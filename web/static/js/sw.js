const CACHE_VERSION = 1;
const CACHE = "cais-static-v" + CACHE_VERSION;

const PRECACHE = [
  "/static/css/styles.css",
  "/static/js/htmx.min.js",
  "/static/js/idiomorph-ext.min.js",
  "/static/js/sse-ext.min.js",
  "/static/js/cais-core.js",
  "/static/js/cais-chat.js",
  "/static/js/html5-qrcode.min.js",
  "/static/js/scan.js",
  "/static/manifest.webmanifest",
  "/static/icons/icon-192.png",
  "/static/icons/icon-512.png",
  "/static/offline.html",
];

self.addEventListener("install", (event) => {
  event.waitUntil(
    caches
      .open(CACHE)
      .then((cache) => cache.addAll(PRECACHE))
      .then(() => self.skipWaiting())
  );
});

self.addEventListener("activate", (event) => {
  event.waitUntil(
    caches
      .keys()
      .then((keys) => Promise.all(keys.filter((k) => k !== CACHE).map((k) => caches.delete(k))))
      .then(() => self.clients.claim())
  );
});

// Network-first for SPA bundles and Tailwind CSS: stable paths like
// /static/build/assets/main.js must pick up vite/tailwind rebuilds without
// requiring a CACHE_VERSION bump or hard refresh.
function networkFirst(request) {
  return fetch(request)
    .then((response) => {
      if (response && response.ok) {
        const copy = response.clone();
        caches.open(CACHE).then((cache) => cache.put(request, copy));
      }
      return response;
    })
    .catch(() => caches.match(request).then((cached) => cached || Response.error()));
}

function cacheFirst(request) {
  return caches.match(request).then((cached) => {
    if (cached) return cached;
    return fetch(request).then((response) => {
      if (response && response.ok) {
        const copy = response.clone();
        caches.open(CACHE).then((cache) => cache.put(request, copy));
      }
      return response;
    });
  });
}

self.addEventListener("fetch", (event) => {
  const { request } = event;
  const url = new URL(request.url);

  if (request.method !== "GET") {
    return;
  }

  if (url.pathname.startsWith("/static/build/") || url.pathname.startsWith("/static/css/")) {
    event.respondWith(networkFirst(request));
    return;
  }

  if (url.pathname.startsWith("/static/")) {
    event.respondWith(cacheFirst(request));
    return;
  }

  if (request.headers.get("accept")?.includes("text/html")) {
    event.respondWith(
      fetch(request).catch(() =>
        caches.match("/static/offline.html").then((cached) => cached || Response.error())
      )
    );
  }
});
