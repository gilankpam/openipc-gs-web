import { useState, useEffect } from 'react';

export function useConnectionStatus(intervalMs = 3000) {
    const [isConnected, setIsConnected] = useState(true);

    useEffect(() => {
        const checkConnection = async () => {
            try {
                const res = await fetch('/api/v1/ping');
                if (res.ok) {
                    setIsConnected(true);
                } else {
                    setIsConnected(false);
                }
            } catch (error) {
                setIsConnected(false);
            }
        };

        // Initial check
        checkConnection();

        const intervalId = setInterval(checkConnection, intervalMs);

        return () => clearInterval(intervalId);
    }, [intervalMs]);

    return isConnected;
}
