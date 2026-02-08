import { useEffect, useRef, useState } from 'react';

type ConnectionState = 'disconnected' | 'connecting' | 'connected' | 'failed';

export function VideoPlayer() {
    const videoRef = useRef<HTMLVideoElement>(null);
    const peerConnectionRef = useRef<RTCPeerConnection | null>(null);
    const [connectionState, setConnectionState] = useState<ConnectionState>('disconnected');

    useEffect(() => {
        let mounted = true;

        const connect = async () => {
            if (!mounted) return;
            setConnectionState('connecting');

            try {
                // Create peer connection
                const pc = new RTCPeerConnection({
                    iceServers: [{ urls: 'stun:stun.l.google.com:19302' }]
                });
                peerConnectionRef.current = pc;

                // Handle incoming tracks
                pc.ontrack = (event) => {
                    if (videoRef.current && event.streams[0]) {
                        videoRef.current.srcObject = event.streams[0];
                    }
                };

                // Handle connection state changes
                pc.oniceconnectionstatechange = () => {
                    if (!mounted) return;
                    switch (pc.iceConnectionState) {
                        case 'connected':
                        case 'completed':
                            setConnectionState('connected');
                            break;
                        case 'failed':
                        case 'closed':
                            setConnectionState('failed');
                            break;
                        case 'disconnected':
                            setConnectionState('disconnected');
                            break;
                    }
                };

                // Add transceiver for receiving video
                pc.addTransceiver('video', { direction: 'recvonly' });

                // Create offer
                const offer = await pc.createOffer();
                await pc.setLocalDescription(offer);

                // Wait for ICE gathering to complete
                await new Promise<void>((resolve) => {
                    if (pc.iceGatheringState === 'complete') {
                        resolve();
                    } else {
                        const checkState = () => {
                            if (pc.iceGatheringState === 'complete') {
                                pc.removeEventListener('icegatheringstatechange', checkState);
                                resolve();
                            }
                        };
                        pc.addEventListener('icegatheringstatechange', checkState);
                    }
                });

                // Send offer to server and get answer
                const response = await fetch('/api/v1/stream/offer', {
                    method: 'POST',
                    headers: { 'Content-Type': 'application/json' },
                    body: JSON.stringify({ offer: pc.localDescription })
                });

                if (!response.ok) {
                    throw new Error(`Server returned ${response.status}`);
                }

                const { answer } = await response.json();
                await pc.setRemoteDescription(answer);

            } catch (error) {
                console.error('WebRTC connection failed:', error);
                if (mounted) {
                    setConnectionState('failed');
                }
            }
        };

        connect();

        return () => {
            mounted = false;
            if (peerConnectionRef.current) {
                peerConnectionRef.current.close();
                peerConnectionRef.current = null;
            }
        };
    }, []);

    const handleRetry = () => {
        if (peerConnectionRef.current) {
            peerConnectionRef.current.close();
            peerConnectionRef.current = null;
        }
        setConnectionState('disconnected');
        // Trigger re-connect by forcing a re-render
        window.location.reload();
    };

    return (
        <div style={{
            width: '100%',
            height: '100vh',
            overflow: 'hidden',
            position: 'relative',
            backgroundColor: '#000'
        }}>
            <video
                ref={videoRef}
                autoPlay
                playsInline
                muted
                style={{
                    width: '100%',
                    height: '100%',
                    objectFit: 'contain'
                }}
            />

            {/* Connection Status Overlay */}
            {connectionState !== 'connected' && (
                <div style={{
                    position: 'absolute',
                    top: 0,
                    left: 0,
                    right: 0,
                    bottom: 0,
                    display: 'flex',
                    flexDirection: 'column',
                    alignItems: 'center',
                    justifyContent: 'center',
                    backgroundColor: 'rgba(0, 0, 0, 0.7)',
                    color: '#fff'
                }}>
                    {connectionState === 'connecting' && (
                        <>
                            <div style={{
                                width: '40px',
                                height: '40px',
                                border: '4px solid #333',
                                borderTopColor: '#fff',
                                borderRadius: '50%',
                                animation: 'spin 1s linear infinite'
                            }} />
                            <p style={{ marginTop: '16px' }}>Connecting to stream...</p>
                        </>
                    )}
                    {connectionState === 'failed' && (
                        <>
                            <p style={{ color: '#ff6b6b', marginBottom: '16px' }}>
                                Connection failed
                            </p>
                            <button
                                onClick={handleRetry}
                                style={{
                                    padding: '8px 24px',
                                    backgroundColor: '#4CAF50',
                                    color: '#fff',
                                    border: 'none',
                                    borderRadius: '4px',
                                    cursor: 'pointer',
                                    fontSize: '16px'
                                }}
                            >
                                Retry
                            </button>
                        </>
                    )}
                    {connectionState === 'disconnected' && (
                        <p>Waiting for stream...</p>
                    )}
                </div>
            )}

            <style>{`
                @keyframes spin {
                    to { transform: rotate(360deg); }
                }
            `}</style>
        </div>
    );
}
