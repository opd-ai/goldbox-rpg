/**
 * Spatial Query Manager - TypeScript implementation with enhanced caching and validation
 * Provides efficient object queries using server-side spatial indexing
 */

import type { Position, Rectangle, GameObject } from '../types/GameTypes';
import type { TypedRPCCall, GetObjectsInRangeParams, GetObjectsInRadiusParams, GetNearestObjectsParams, SpatialQueryResult } from '../types/RPCTypes';
import { logger } from './Logger';
import { createErrorHandler } from './ErrorHandler';

interface CacheEntry {
  readonly data: readonly GameObject[];
  readonly timestamp: number;
  readonly ttl: number;
}

interface CacheStats {
  readonly hits: number;
  readonly misses: number;
  readonly size: number;
  readonly hitRate: number;
  readonly averageTTL: number;
}

interface SpatialQueryManagerConfig {
  readonly defaultCacheTimeout: number;
  readonly maxCacheSize: number;
  readonly cleanupInterval: number;
  readonly adaptiveCaching: boolean;
}

export class SpatialQueryManager {
  private readonly rpc: TypedRPCCall;
  private readonly cache = new Map<string, CacheEntry>();
  private readonly config: SpatialQueryManagerConfig;
  private readonly errorHandler = createErrorHandler({ component: 'SpatialQueryManager' });
  private readonly logger = logger.createChildLogger('SpatialQueryManager');
  
  // Cache statistics
  private cacheHits = 0;
  private cacheMisses = 0;
  private cleanupTimer?: ReturnType<typeof setTimeout> | undefined;

  // Adaptive cache timeouts based on object type
  private readonly cacheTimeouts = new Map<string, number>([
    ['static_objects', 30000],    // 30 seconds - static objects change rarely
    ['dynamic_objects', 1000],    // 1 second - dynamic objects change frequently
    ['players', 500],             // 0.5 seconds - players move frequently
    ['items', 5000],              // 5 seconds - items change moderately
    ['npcs', 2000],               // 2 seconds - NPCs change regularly
    ['decorations', 60000],       // 1 minute - decorations never change
  ]);

  constructor(
    rpcCall: TypedRPCCall,
    config: Partial<SpatialQueryManagerConfig> = {}
  ) {
    this.rpc = rpcCall;
    this.config = {
      defaultCacheTimeout: 1000,
      maxCacheSize: 1000,
      cleanupInterval: 30000,
      adaptiveCaching: true,
      ...config
    };

    this.startCleanupTimer();
    this.logger.info('SpatialQueryManager initialized', { config: this.config });
  }

  /**
   * Get all objects within a rectangular area using server-side spatial indexing
   */
  async getObjectsInRange(
    rect: Rectangle,
    sessionId: string,
    objectType: string = 'dynamic_objects'
  ): Promise<readonly GameObject[]> {
    return this.errorHandler.wrapAsync(
      async () => {
        this.validateRectangle(rect);
        this.validateSessionId(sessionId);

        const cacheKey = this.generateRangeCacheKey(rect, objectType);
        
        // Check cache first
        const cached = this.getCachedResult(cacheKey);
        if (cached) {
          this.cacheHits++;
          this.logger.debug('Cache hit for range query', { rect, objectType });
          return cached;
        }

        this.cacheMisses++;
        this.logger.debug('Cache miss for range query', { rect, objectType });

        // Query server
        const params: GetObjectsInRangeParams = {
          session_id: sessionId,
          min_x: rect.x,
          min_y: rect.y,
          max_x: rect.x + rect.width,
          max_y: rect.y + rect.height
        };

        const result = await this.rpc('getObjectsInRange', params);
        const objects = this.validateAndTransformResult(result);
        
        // Cache the result
        this.setCachedResult(cacheKey, objects, objectType);
        
        return objects;
      },
      'getObjectsInRange',
      {
        userMessage: 'Failed to query objects in range',
        metadata: { rect, sessionId, objectType }
      }
    )();
  }

  /**
   * Get all objects within a circular area using server-side spatial indexing
   */
  async getObjectsInRadius(
    center: Position,
    radius: number,
    sessionId: string,
    objectType: string = 'dynamic_objects'
  ): Promise<readonly GameObject[]> {
    return this.errorHandler.wrapAsync(
      async () => {
        this.validatePosition(center);
        this.validateRadius(radius);
        this.validateSessionId(sessionId);

        const cacheKey = this.generateRadiusCacheKey(center, radius, objectType);
        
        // Check cache first
        const cached = this.getCachedResult(cacheKey);
        if (cached) {
          this.cacheHits++;
          this.logger.debug('Cache hit for radius query', { center, radius, objectType });
          return cached;
        }

        this.cacheMisses++;
        this.logger.debug('Cache miss for radius query', { center, radius, objectType });

        // Query server
        const params: GetObjectsInRadiusParams = {
          session_id: sessionId,
          center_x: center.x,
          center_y: center.y,
          radius
        };

        const result = await this.rpc('getObjectsInRadius', params);
        const objects = this.validateAndTransformResult(result);
        
        // Cache the result
        this.setCachedResult(cacheKey, objects, objectType);
        
        return objects;
      },
      'getObjectsInRadius',
      {
        userMessage: 'Failed to query objects in radius',
        metadata: { center, radius, sessionId, objectType }
      }
    )();
  }

  /**
   * Get the k nearest objects to a position using server-side spatial indexing
   */
  async getNearestObjects(
    center: Position,
    k: number,
    sessionId: string,
    objectType: string = 'dynamic_objects'
  ): Promise<readonly GameObject[]> {
    return this.errorHandler.wrapAsync(
      async () => {
        this.validatePosition(center);
        this.validateNearestCount(k);
        this.validateSessionId(sessionId);

        const cacheKey = this.generateNearestCacheKey(center, k, objectType);
        
        // Check cache first
        const cached = this.getCachedResult(cacheKey);
        if (cached) {
          this.cacheHits++;
          this.logger.debug('Cache hit for nearest query', { center, k, objectType });
          return cached;
        }

        this.cacheMisses++;
        this.logger.debug('Cache miss for nearest query', { center, k, objectType });

        // Query server
        const params: GetNearestObjectsParams = {
          session_id: sessionId,
          center_x: center.x,
          center_y: center.y,
          k
        };

        const result = await this.rpc('getNearestObjects', params);
        const objects = this.validateAndTransformResult(result);
        
        // Cache the result
        this.setCachedResult(cacheKey, objects, objectType);
        
        return objects;
      },
      'getNearestObjects',
      {
        userMessage: 'Failed to query nearest objects',
        metadata: { center, k, sessionId, objectType }
      }
    )();
  }

  /**
   * Clear the query cache
   */
  clearCache(objectType?: string): void {
    if (objectType) {
      // Clear only caches for specific object type
      const keysToDelete: string[] = [];
      for (const key of Array.from(this.cache.keys())) {
        if (key.includes(`type:${objectType}`)) {
          keysToDelete.push(key);
        }
      }
      
      keysToDelete.forEach(key => this.cache.delete(key));
      this.logger.debug(`Cleared cache for object type: ${objectType}`, { 
        clearedEntries: keysToDelete.length 
      });
    } else {
      // Clear all caches
      const size = this.cache.size;
      this.cache.clear();
      this.logger.debug('Cleared all cache entries', { clearedEntries: size });
    }
  }

  /**
   * Get cache statistics for monitoring and optimization
   */
  getCacheStats(): CacheStats {
    const totalRequests = this.cacheHits + this.cacheMisses;
    const hitRate = totalRequests > 0 ? this.cacheHits / totalRequests : 0;
    
    // Calculate average TTL
    let totalTTL = 0;
    for (const entry of Array.from(this.cache.values())) {
      totalTTL += entry.ttl;
    }
    const averageTTL = this.cache.size > 0 ? totalTTL / this.cache.size : 0;

    return {
      hits: this.cacheHits,
      misses: this.cacheMisses,
      size: this.cache.size,
      hitRate,
      averageTTL
    };
  }

  /**
   * Configure cache timeouts for specific object types
   */
  configureCacheTimeouts(timeoutConfig: Record<string, number>): void {
    for (const [objectType, timeout] of Object.entries(timeoutConfig)) {
      this.cacheTimeouts.set(objectType, timeout);
    }
    
    this.logger.info('Updated cache timeouts', { timeoutConfig });
  }

  /**
   * Cleanup and destroy the spatial query manager
   */
  destroy(): void {
    if (this.cleanupTimer) {
      clearInterval(this.cleanupTimer);
      this.cleanupTimer = undefined;
    }
    
    this.clearCache();
    this.logger.info('SpatialQueryManager destroyed');
  }

  // Private helper methods

  private getCachedResult(key: string): readonly GameObject[] | null {
    const entry = this.cache.get(key);
    if (!entry) {
      return null;
    }

    // Check if expired
    if (Date.now() > entry.timestamp + entry.ttl) {
      this.cache.delete(key);
      return null;
    }

    return entry.data;
  }

  private setCachedResult(
    key: string,
    data: readonly GameObject[],
    objectType: string
  ): void {
    // Enforce cache size limit
    if (this.cache.size >= this.config.maxCacheSize) {
      this.evictOldestEntries();
    }

    const ttl = this.getCacheTimeout(objectType);
    const entry: CacheEntry = {
      data,
      timestamp: Date.now(),
      ttl
    };

    this.cache.set(key, entry);
  }

  private getCacheTimeout(objectType: string): number {
    if (this.config.adaptiveCaching && this.cacheTimeouts.has(objectType)) {
      return this.cacheTimeouts.get(objectType)!;
    }
    return this.config.defaultCacheTimeout;
  }

  private evictOldestEntries(): void {
    // Remove 20% of oldest entries
    const entriesToRemove = Math.floor(this.cache.size * 0.2);
    const entries = Array.from(this.cache.entries()).sort((a, b) => 
      a[1].timestamp - b[1].timestamp
    );

    for (let i = 0; i < entriesToRemove && i < entries.length; i++) {
      this.cache.delete(entries[i][0]);
    }
  }

  private startCleanupTimer(): void {
    this.cleanupTimer = setInterval(() => {
      this.cleanupExpiredEntries();
    }, this.config.cleanupInterval);
  }

  private cleanupExpiredEntries(): void {
    const now = Date.now();
    const keysToDelete: string[] = [];

    for (const [key, entry] of Array.from(this.cache.entries())) {
      if (now > entry.timestamp + entry.ttl) {
        keysToDelete.push(key);
      }
    }

    keysToDelete.forEach(key => this.cache.delete(key));
    
    if (keysToDelete.length > 0) {
      this.logger.debug('Cleaned up expired cache entries', { 
        cleanedEntries: keysToDelete.length,
        remainingEntries: this.cache.size
      });
    }
  }

  // Cache key generation methods
  private generateRangeCacheKey(rect: Rectangle, objectType: string): string {
    return `range:${rect.x},${rect.y},${rect.width},${rect.height}:type:${objectType}`;
  }

  private generateRadiusCacheKey(center: Position, radius: number, objectType: string): string {
    return `radius:${center.x},${center.y},${radius}:type:${objectType}`;
  }

  private generateNearestCacheKey(center: Position, k: number, objectType: string): string {
    return `nearest:${center.x},${center.y},${k}:type:${objectType}`;
  }

  // Validation methods
  private validateRectangle(rect: Rectangle): void {
    if (!rect || typeof rect.x !== 'number' || typeof rect.y !== 'number' ||
        typeof rect.width !== 'number' || typeof rect.height !== 'number') {
      throw new Error('Invalid rectangle parameters');
    }
    if (rect.width <= 0 || rect.height <= 0) {
      throw new Error('Rectangle dimensions must be positive');
    }
  }

  private validatePosition(position: Position): void {
    if (!position || typeof position.x !== 'number' || typeof position.y !== 'number') {
      throw new Error('Invalid position parameters');
    }
  }

  private validateRadius(radius: number): void {
    if (typeof radius !== 'number' || radius <= 0) {
      throw new Error('Radius must be a positive number');
    }
  }

  private validateNearestCount(k: number): void {
    if (typeof k !== 'number' || k <= 0 || !Number.isInteger(k)) {
      throw new Error('k must be a positive integer');
    }
  }

  private validateSessionId(sessionId: string): void {
    if (typeof sessionId !== 'string' || sessionId.trim().length === 0) {
      throw new Error('Invalid session ID');
    }
  }

  private validateAndTransformResult(result: SpatialQueryResult): readonly GameObject[] {
    if (!result || !Array.isArray(result.objects)) {
      throw new Error('Invalid spatial query result format');
    }

    // Transform and validate each object
    return result.objects.map((obj, index) => {
      if (!obj || typeof obj !== 'object') {
        throw new Error(`Invalid object at index ${index}`);
      }

      // Basic validation and transformation
      const gameObject = obj as GameObject;
      if (!gameObject.id || !gameObject.position) {
        throw new Error(`Invalid game object structure at index ${index}`);
      }

      return gameObject;
    });
  }
}
