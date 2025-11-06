import {useCallback} from 'react';
import {useAuth} from '../AuthContext';

/**
 * Custom hook that wraps fetch with automatic JWT refresh on 401 errors
 * Usage: const authFetch = useAuthFetch();
 *        const response = await authFetch('/api/endpoint', options);
 */
export function useAuthFetch() {
    const { jwt, refreshJwt, clearAuth } = useAuth();

    return useCallback(async (url, options = {}) => {
        // Add Authorization header if not already present
        const headers = {
            ...options.headers,
        };

        if (jwt && !headers.Authorization) {
            headers.Authorization = `Bearer ${jwt}`;
        }

        const fetchOptions = {
            ...options,
            headers,
        };

        // First attempt
        let response = await fetch(url, fetchOptions);

        // If 401 Unauthorized, try to refresh token and retry
        if (response.status === 401) {
            console.log("Token expired, attempting refresh...");

            const newJwt = await refreshJwt();

            if (newJwt) {
                // Retry with new token
                headers.Authorization = `Bearer ${newJwt}`;
                response = await fetch(url, {...fetchOptions, headers});

                console.log("Request retried with refreshed token");
            } else {
                // Refresh failed, clear auth and redirect to login
                console.error("Token refresh failed, clearing auth");
                clearAuth();
                throw new Error("Session expired. Please log in again.");
            }
        }

        return response;
    }, [jwt, refreshJwt, clearAuth]);
}

/**
 * Alternative: Create a configured fetch instance
 * This can be used directly without a hook
 */
export function createAuthFetch(getAuth, refreshJwt, clearAuth) {
    return async (url, options = {}) => {
        const { jwt } = getAuth();

        const headers = {
            ...options.headers,
        };

        if (jwt && !headers.Authorization) {
            headers.Authorization = `Bearer ${jwt}`;
        }

        const fetchOptions = {
            ...options,
            headers,
        };

        let response = await fetch(url, fetchOptions);

        if (response.status === 401) {
            const newJwt = await refreshJwt();

            if (newJwt) {
                headers.Authorization = `Bearer ${newJwt}`;
                response = await fetch(url, { ...fetchOptions, headers });
            } else {
                clearAuth();
                throw new Error("Session expired. Please log in again.");
            }
        }

        return response;
    };
}