const BASE_PATH = process.env.REACT_APP_BASE_PATH || "";

function urlBase64ToUint8Array(base64String) {
    const padding = '='.repeat((4 - base64String.length % 4) % 4);
    const base64 = (base64String + padding)
        .replace(/-/g, '+')
        .replace(/_/g, '/');

    const rawData = window.atob(base64);
    const outputArray = new Uint8Array(rawData.length);

    for (let i = 0; i < rawData.length; ++i) {
        outputArray[i] = rawData.charCodeAt(i);
    }
    return outputArray;
}

// Convert ArrayBuffer to base64 string
function arrayBufferToBase64(buffer) {
    const bytes = new Uint8Array(buffer);
    let binary = '';
    for (let i = 0; i < bytes.byteLength; i++) {
        binary += String.fromCharCode(bytes[i]);
    }
    return window.btoa(binary);
}

export async function registerServiceWorker() {
    if (!('serviceWorker' in navigator)) {
        console.warn('Service workers not supported');
        return null;
    }
    try {
        const registration = await navigator.serviceWorker.register('/service-worker.js');
        console.log('✅ Service worker registered:', registration);
        return registration;
    } catch (error) {
        console.error('Failed to register service worker:', error);
        return null;
    }
}

export async function requestNotificationPermission() {
    if (!('Notification' in window)) {
        console.warn('Notifications not supported');
        return false;
    }

    if (Notification.permission === 'granted') {
        return true;
    }

    if (Notification.permission === 'denied') {
        return false;
    }

    const permission = await Notification.requestPermission();
    return permission === 'granted';
}

export async function getVapidPublicKey() {
    try {
        const response = await fetch(`${BASE_PATH}/api/push/vapid-public-key`);
        if (!response.ok) {
            throw new Error('Failed to fetch VAPID public key');
        }
        const data = await response.json();
        return data.publicKey;
    } catch (error) {
        console.error('Error fetching VAPID key:', error);
        return null;
    }
}

export async function subscribeToPushNotifications(token) {
    if (!token) {
        console.error('No auth token provided');
        return false;
    }

    const registration = await registerServiceWorker();
    if (!registration) {
        return false;
    }

    const hasPermission = await requestNotificationPermission();
    if (!hasPermission) {
        console.warn('Notification permission denied');
        return false;
    }

    try {
        // Get VAPID public key from server
        const vapidPublicKey = await getVapidPublicKey();
        if (!vapidPublicKey) {
            console.error('No VAPID public key available');
            return false;
        }

        // Subscribe to push notifications
        const subscription = await registration.pushManager.subscribe({
            userVisibleOnly: true,
            applicationServerKey: urlBase64ToUint8Array(vapidPublicKey)
        });

        console.log('✅ Push subscription:', subscription);

        // Send subscription to server
        const response = await fetch(`${BASE_PATH}/api/push/subscribe`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                endpoint: subscription.endpoint,
                keys: {
                    p256dh: arrayBufferToBase64(subscription.getKey('p256dh')),
                    auth: arrayBufferToBase64(subscription.getKey('auth'))
                }
            })
        });

        if (!response.ok) {
            throw new Error('Failed to save subscription on server');
        }

        console.log('✅ Push subscription saved on server');
        return true;
    } catch (error) {
        console.error('Failed to subscribe to push notifications:', error);
        return false;
    }
}

export async function unsubscribeFromPushNotifications(token) {
    if (!token) {
        return false;
    }

    try {
        const registration = await navigator.serviceWorker.getRegistration();
        if (!registration) {
            return false;
        }

        const subscription = await registration.pushManager.getSubscription();
        if (subscription) {
            await subscription.unsubscribe();
        }

        // Remove subscription from server
        await fetch(`${BASE_PATH}/api/push/unsubscribe`, {
            method: 'POST',
            headers: {
                'Authorization': `Bearer ${token}`
            }
        });

        console.log('✅ Push subscription removed');
        return true;
    } catch (error) {
        console.error('Failed to unsubscribe from push notifications:', error);
        return false;
    }
}

export async function inviteUserToRoom(token, roomId, invitedUsername) {
    try {
        const response = await fetch(`${BASE_PATH}/api/rooms/${roomId}/invite`, {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
                'Authorization': `Bearer ${token}`
            },
            body: JSON.stringify({
                invited_username: invitedUsername
            })
        });

        if (!response.ok) {
            const errorText = await response.text();
            throw new Error(errorText || 'Failed to send invitation');
        }

        return true;
    } catch (error) {
        console.error('Failed to invite user:', error);
        throw error;
    }
}