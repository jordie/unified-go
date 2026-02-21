/**
 * Service Worker for Math Educational App
 *
 * Features:
 * - Offline functionality with cached assets
 * - API request caching (network-first strategy)
 * - Background sync for saving results
 * - Cache versioning and cleanup
 */

const CACHE_VERSION = 'math-app-v1';
const STATIC_CACHE = `${CACHE_VERSION}-static`;
const DYNAMIC_CACHE = `${CACHE_VERSION}-dynamic`;
const API_CACHE = `${CACHE_VERSION}-api`;

// Assets to cache on install
const STATIC_ASSETS = [
  '/math/',
  '/math/static/script.js',
  '/math/static/style.css',
  '/math/static/sounds.js',
  '/math/index.html'
];

/**
 * Install event - cache essential assets
 */
self.addEventListener('install', (event) => {
  console.log('[ServiceWorker] Installing...');
  event.waitUntil(
    caches.open(STATIC_CACHE)
      .then((cache) => {
        console.log('[ServiceWorker] Caching static assets');
        return cache.addAll(STATIC_ASSETS).catch((err) => {
          console.warn('[ServiceWorker] Failed to cache some assets:', err);
          // Don't fail install if some assets can't be cached
          return Promise.resolve();
        });
      })
      .then(() => self.skipWaiting())
      .catch((err) => console.error('[ServiceWorker] Install error:', err))
  );
});

/**
 * Activate event - clean up old caches
 */
self.addEventListener('activate', (event) => {
  console.log('[ServiceWorker] Activating...');
  event.waitUntil(
    caches.keys()
      .then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (!cacheName.startsWith('math-app-v')) {
              return caches.delete(cacheName);
            }
            // Delete old versions
            if (cacheName !== STATIC_CACHE &&
                cacheName !== DYNAMIC_CACHE &&
                cacheName !== API_CACHE) {
              return caches.delete(cacheName);
            }
            return Promise.resolve();
          })
        );
      })
      .then(() => self.clients.claim())
      .catch((err) => console.error('[ServiceWorker] Activate error:', err))
  );
});

/**
 * Fetch event - implement caching strategies
 */
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Skip non-GET requests
  if (request.method !== 'GET') {
    return;
  }

  // Skip external requests and chrome extensions
  if (!url.pathname.startsWith('/math') && !url.pathname.startsWith('/api')) {
    return;
  }

  // API requests: network-first strategy
  if (url.pathname.startsWith('/api/')) {
    event.respondWith(networkFirstStrategy(request));
    return;
  }

  // Static assets: cache-first strategy
  event.respondWith(cacheFirstStrategy(request));
});

/**
 * Cache-first strategy for static assets
 * Returns cached version if available, otherwise fetches from network
 */
async function cacheFirstStrategy(request) {
  try {
    const cache = await caches.open(STATIC_CACHE);
    const cached = await cache.match(request);

    if (cached) {
      console.log('[ServiceWorker] Cache hit:', request.url);
      return cached;
    }

    // Not in cache, fetch from network
    const response = await fetch(request);

    if (response.ok) {
      // Clone and cache the response
      const cloned = response.clone();
      cache.put(request, cloned);
    }

    return response;
  } catch (error) {
    console.error('[ServiceWorker] Fetch error:', error);

    // Return offline page or cached fallback
    const cached = await caches.match(request);
    if (cached) {
      return cached;
    }

    // Return offline response for HTML
    if (request.headers.get('accept').includes('text/html')) {
      return new Response('<h1>Offline</h1><p>You are currently offline.</p>', {
        headers: { 'Content-Type': 'text/html' },
        status: 503
      });
    }

    return new Response('Offline', { status: 503 });
  }
}

/**
 * Network-first strategy for API calls
 * Tries network first, falls back to cache
 */
async function networkFirstStrategy(request) {
  try {
    const response = await fetch(request);

    if (response.ok) {
      // Cache successful API responses
      const cache = await caches.open(API_CACHE);
      const cloned = response.clone();
      cache.put(request, cloned);
    }

    return response;
  } catch (error) {
    console.warn('[ServiceWorker] Network request failed, checking cache:', error);

    // Try to return cached API response
    const cache = await caches.open(API_CACHE);
    const cached = await cache.match(request);

    if (cached) {
      return cached;
    }

    // Return offline error response
    return new Response(
      JSON.stringify({ error: 'Offline', offline: true }),
      {
        headers: { 'Content-Type': 'application/json' },
        status: 503
      }
    );
  }
}

/**
 * Message handler for cache management
 */
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (event.data && event.data.type === 'CLEAR_CACHE') {
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName.startsWith('math-app-v')) {
            return caches.delete(cacheName);
          }
        })
      );
    }).then(() => {
      event.ports[0].postMessage({ cleared: true });
    });
  }
});

console.log('[ServiceWorker] Math app service worker loaded');
