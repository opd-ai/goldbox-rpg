/**
 * Spatial Query Manager - TypeScript implementation with enhanced caching and validation
 * Provides efficient object queries using server-side spatial indexing
 */
import type { Position, Rectangle, GameObject } from '../types/GameTypes';
import type { TypedRPCCall } from '../types/RPCTypes';
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
export declare class SpatialQueryManager {
    private readonly rpc;
    private readonly cache;
    private readonly config;
    private readonly errorHandler;
    private readonly logger;
    private cacheHits;
    private cacheMisses;
    private cleanupTimer?;
    private readonly cacheTimeouts;
    constructor(rpcCall: TypedRPCCall, config?: Partial<SpatialQueryManagerConfig>);
    /**
     * Get all objects within a rectangular area using server-side spatial indexing
     */
    getObjectsInRange(rect: Rectangle, sessionId: string, objectType?: string): Promise<readonly GameObject[]>;
    /**
     * Get all objects within a circular area using server-side spatial indexing
     */
    getObjectsInRadius(center: Position, radius: number, sessionId: string, objectType?: string): Promise<readonly GameObject[]>;
    /**
     * Get the k nearest objects to a position using server-side spatial indexing
     */
    getNearestObjects(center: Position, k: number, sessionId: string, objectType?: string): Promise<readonly GameObject[]>;
    /**
     * Clear the query cache
     */
    clearCache(objectType?: string): void;
    /**
     * Get cache statistics for monitoring and optimization
     */
    getCacheStats(): CacheStats;
    /**
     * Configure cache timeouts for specific object types
     */
    configureCacheTimeouts(timeoutConfig: Record<string, number>): void;
    /**
     * Cleanup and destroy the spatial query manager
     */
    destroy(): void;
    private getCachedResult;
    private setCachedResult;
    private getCacheTimeout;
    private evictOldestEntries;
    private startCleanupTimer;
    private cleanupExpiredEntries;
    private generateRangeCacheKey;
    private generateRadiusCacheKey;
    private generateNearestCacheKey;
    private validateRectangle;
    private validatePosition;
    private validateRadius;
    private validateNearestCount;
    private validateSessionId;
    private validateAndTransformResult;
}
export {};
//# sourceMappingURL=SpatialQueryManager.d.ts.map