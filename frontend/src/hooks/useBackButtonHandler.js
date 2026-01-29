import { useEffect } from 'react';
import { useLocation } from 'react-router-dom';

export function useBackButtonHandler(options = {}) {
    const {
        isExitPage = false,
        preventBack = false,
        onBackPress = null,
        confirmMessage = 'Хотите выйти?'
    } = options;

    const location = useLocation();

    useEffect(() => {
        const handlePopState = (e) => {
            e.preventDefault();

            if (preventBack) {
                // Prevent going back (stay on current page)
                window.history.pushState(null, '', location.pathname);

                if (onBackPress) {
                    onBackPress();
                }
                return;
            }

            if (isExitPage) {
                // This is the exit page (home/create room)
                const shouldExit = window.confirm(confirmMessage);

                if (shouldExit) {
                    // Close the app
                    closeApp();
                } else {
                    // Push state again to stay on page
                    window.history.pushState(null, '', location.pathname);
                }
            }
        };

        // Add initial state
        window.history.pushState(null, '', location.pathname);
        window.addEventListener('popstate', handlePopState);

        return () => {
            window.removeEventListener('popstate', handlePopState);
        };
    }, [isExitPage, preventBack, onBackPress, location.pathname, confirmMessage]);
}

function closeApp() {
    // Try different methods to close the app

    // Method 1: Standard window.close()
    window.close();

    // Method 2: For Android WebView
    if (window.Android && window.Android.closeApp) {
        window.Android.closeApp();
    }

    // Method 3: For iOS WebView
    if (window.webkit && window.webkit.messageHandlers && window.webkit.messageHandlers.closeApp) {
        window.webkit.messageHandlers.closeApp.postMessage('close');
    }

    // Method 4: Navigate to blank page (fallback)
    setTimeout(() => {
        if (!window.closed) {
            window.location.href = 'about:blank';
        }
    }, 100);
}