/* eslint-disable no-restricted-globals */

// Install event - cache static assets if needed
self.addEventListener('install', (event) => {
    console.log('Service Worker installing...');
    self.skipWaiting();
});

// Activate event - clean up old caches
self.addEventListener('activate', (event) => {
    console.log('Service Worker activating...');
    event.waitUntil(self.clients.claim());
});

// Push event - handle incoming push notifications
self.addEventListener('push', (event) => {
    console.log('Push notification received:', event);

    let notificationData = {
        title: 'New Notification',
        body: 'You have a new message',
        icon: '/icon-192x192.png',
        badge: '/badge-72x72.png',
        tag: 'default-notification',
        requireInteraction: false,
        data: {
            url: '/'
        }
    };

    // Parse push data if available
    if (event.data) {
        try {
            const payload = event.data.json();
            console.log('Push payload:', payload);

            notificationData = {
                title: payload.title || notificationData.title,
                body: payload.body || payload.message || notificationData.body,
                icon: payload.icon || notificationData.icon,
                badge: payload.badge || notificationData.badge,
                tag: payload.tag || notificationData.tag,
                requireInteraction: payload.requireInteraction || false,
                data: {
                    url: payload.url || payload.data?.url || '/',
                    roomId: payload.roomId || payload.data?.roomId,
                    type: payload.type || payload.data?.type,
                    ...payload.data
                },
                actions: payload.actions || []
            };

            // Add image if provided
            if (payload.image) {
                notificationData.image = payload.image;
            }

            // Add vibration pattern if provided
            if (payload.vibrate) {
                notificationData.vibrate = payload.vibrate;
            }
        } catch (error) {
            console.error('Error parsing push data:', error);
            // Use text content if JSON parsing fails
            notificationData.body = event.data.text();
        }
    }

    // Show the notification
    event.waitUntil(
        self.registration.showNotification(notificationData.title, {
            body: notificationData.body,
            icon: notificationData.icon,
            badge: notificationData.badge,
            tag: notificationData.tag,
            requireInteraction: notificationData.requireInteraction,
            data: notificationData.data,
            actions: notificationData.actions,
            image: notificationData.image,
            vibrate: notificationData.vibrate
        })
    );
});

self.addEventListener('notificationclick', (event) => {
    console.log('Notification data:', event.notification.data);

    event.notification.close();

    // Determine URL to open based on notification data
    let urlToOpen = '/';

    if (event.notification.data) {
        const data = event.notification.data;

        // If roomId is present, navigate to join room
        if (data.roomId) {
            urlToOpen = `/join/${data.roomId}`;
            console.log('Opening room:', data.roomId);
        }
        // Otherwise use the url field if provided
        else if (data.url) {
            urlToOpen = data.url;
        }
    }

    // Open or focus the app
    event.waitUntil(
        self.clients.matchAll({
            type: 'window',
            includeUncontrolled: true
        }).then((clientList) => {
            const targetUrl = new URL(urlToOpen, self.location.origin);

            // Check if there's already a window open
            for (let i = 0; i < clientList.length; i++) {
                const client = clientList[i];
                const clientUrl = new URL(client.url);

                // If we find a window from the same origin, navigate and focus it
                if (clientUrl.origin === targetUrl.origin) {
                    console.log('Focusing existing window and navigating to:', targetUrl.href);
                    return client.navigate(targetUrl.href).then(client => client.focus());
                }
            }

            // If no window is open, open a new one
            if (self.clients.openWindow) {
                console.log('Opening new window:', targetUrl.href);
                return self.clients.openWindow(targetUrl.href);
            }
        }).catch(error => {
            console.error('Error handling notification click:', error);
        })
    );
});

self.addEventListener('notificationclose', (event) => {
    console.log('Notification closed:', event);
});

self.addEventListener('message', (event) => {
    console.log('Service Worker received message:', event.data);

    if (event.data && event.data.type === 'SKIP_WAITING') {
        self.skipWaiting();
    }

    // Handle other message types as needed
});

// Fetch event - you can add caching strategies here if needed
self.addEventListener('fetch', (event) => {
    // Pass through all fetch requests without caching for now
    // You can implement caching strategies here if needed
    event.respondWith(fetch(event.request));
});