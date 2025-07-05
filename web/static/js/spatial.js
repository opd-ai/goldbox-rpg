/**
 * Spatial Query Manager - Provides efficient object queries using server-side spatial indexing
 * Replaces inefficient client-side object iteration with optimized server queries
 */
class SpatialQueryManager {
    constructor(rpcClient) {
        this.rpc = rpcClient;
        this.cache = new Map();
        this.cacheTimeout = 1000; // 1 second cache timeout
    }

    /**
     * Get all objects within a rectangular area using server-side spatial indexing
     * @param {Object} rect - Rectangle parameters
     * @param {number} rect.minX - Minimum X coordinate
     * @param {number} rect.minY - Minimum Y coordinate  
     * @param {number} rect.maxX - Maximum X coordinate
     * @param {number} rect.maxY - Maximum Y coordinate
     * @param {string} sessionId - Player session ID
     * @returns {Promise<Array>} Array of objects within the range
     */
    async getObjectsInRange(rect, sessionId) {
        const cacheKey = `range_${rect.minX}_${rect.minY}_${rect.maxX}_${rect.maxY}`;
        
        // Check cache first
        const cached = this.cache.get(cacheKey);
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getObjectsInRange', {
                session_id: sessionId,
                min_x: rect.minX,
                min_y: rect.minY,
                max_x: rect.maxX,
                max_y: rect.maxY
            });

            if (result.success) {
                // Cache the result
                this.cache.set(cacheKey, {
                    objects: result.objects,
                    timestamp: Date.now()
                });
                
                return result.objects;
            } else {
                console.error('Range query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Range query error:', error);
            return [];
        }
    }

    /**
     * Get all objects within a circular area using server-side spatial indexing
     * @param {Object} center - Center position
     * @param {number} center.x - X coordinate of center
     * @param {number} center.y - Y coordinate of center
     * @param {number} radius - Search radius
     * @param {string} sessionId - Player session ID
     * @returns {Promise<Array>} Array of objects within the radius
     */
    async getObjectsInRadius(center, radius, sessionId) {
        const cacheKey = `radius_${center.x}_${center.y}_${radius}`;
        
        // Check cache first
        const cached = this.cache.get(cacheKey);
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getObjectsInRadius', {
                session_id: sessionId,
                center_x: center.x,
                center_y: center.y,
                radius: radius
            });

            if (result.success) {
                // Cache the result
                this.cache.set(cacheKey, {
                    objects: result.objects,
                    timestamp: Date.now()
                });
                
                return result.objects;
            } else {
                console.error('Radius query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Radius query error:', error);
            return [];
        }
    }

    /**
     * Get the k nearest objects to a position using server-side spatial indexing
     * @param {Object} center - Center position
     * @param {number} center.x - X coordinate of center
     * @param {number} center.y - Y coordinate of center
     * @param {number} k - Number of nearest objects to return
     * @param {string} sessionId - Player session ID
     * @returns {Promise<Array>} Array of k nearest objects
     */
    async getNearestObjects(center, k, sessionId) {
        const cacheKey = `nearest_${center.x}_${center.y}_${k}`;
        
        // Check cache first
        const cached = this.cache.get(cacheKey);
        if (cached && (Date.now() - cached.timestamp) < this.cacheTimeout) {
            return cached.objects;
        }

        try {
            const result = await this.rpc.call('getNearestObjects', {
                session_id: sessionId,
                center_x: center.x,
                center_y: center.y,
                k: k
            });

            if (result.success) {
                // Cache the result
                this.cache.set(cacheKey, {
                    objects: result.objects,
                    timestamp: Date.now()
                });
                
                return result.objects;
            } else {
                console.error('Nearest objects query failed:', result.error);
                return [];
            }
        } catch (error) {
            console.error('Nearest objects query error:', error);
            return [];
        }
    }

    /**
     * Clear the query cache (call when game state changes significantly)
     */
    clearCache() {
        this.cache.clear();
    }

    /**
     * Remove expired entries from cache
     */
    cleanupCache() {
        const now = Date.now();
        for (const [key, value] of this.cache.entries()) {
            if (now - value.timestamp > this.cacheTimeout) {
                this.cache.delete(key);
            }
        }
    }
}

// Export for use in other modules
if (typeof module !== 'undefined' && module.exports) {
    module.exports = SpatialQueryManager;
} else if (typeof window !== 'undefined') {
    window.SpatialQueryManager = SpatialQueryManager;
}
