interface VideoPlayerProps {
    streamUrl: string;
}

export function VideoPlayer({ streamUrl }: VideoPlayerProps) {
    return (
        <div style={{ width: '100%', height: '100vh', overflow: 'hidden' }}>
            <iframe
                src={streamUrl}
                title="Live Stream"
                style={{ border: 0, width: '100%', height: '100%', display: 'block' }}
                allowFullScreen
            />
        </div>
    );
}
