import { useState, useEffect, useRef } from 'react';
import { fetchWithTimeout } from '../utils/api';

export function useConnectionStatus(intervalMs = 3000) {
    const [isConnected, setIsConnected] = useState(true);
    const timeoutRef = useRef<number | null>(null);

    useEffect(() => {
        let isMounted = true;

        const checkConnection = async () => {
            try {
                const res = await fetchWithTimeout('/api/v1/ping', { timeout: 2000 });
                if (isMounted) {
                    setIsConnected(res.ok);
                }
            } catch (error) {
                if (isMounted) {
                    setIsConnected(false);
                }
            } finally {
                if (isMounted) {
                    timeoutRef.current = setTimeout(checkConnection, intervalMs);
                }
            }
        };

        // Initial check
        checkConnection();

        return () => {
            isMounted = false;
            if (timeoutRef.current) {
                clearTimeout(timeoutRef.current);
            }
        };
    }, [intervalMs]);

    return isConnected;
}
