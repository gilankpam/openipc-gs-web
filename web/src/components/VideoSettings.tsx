import { useState, useEffect } from 'react';
import { Stack, NumberInput, Select, Loader, Text } from '@mantine/core';
import type { VideoSettings as VideoSettingsType } from '../types';
import { fetchWithTimeout } from '../utils/api';

export function VideoSettings() {
    const [settings, setSettings] = useState<VideoSettingsType | null>(null);
    const [connectionError, setConnectionError] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const videoRes = await fetchWithTimeout('/api/v1/video');
                if (!videoRes.ok) throw new Error('Network response was not ok');
                const videoData = await videoRes.json();

                setSettings(videoData);
                setConnectionError(false);
            } catch (err) {
                console.warn('Failed to fetch video settings', err);
                setConnectionError(true);
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const saveSettings = (newSettings: VideoSettingsType) => {
        setSaving(true);
        fetchWithTimeout('/api/v1/video', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newSettings),
        })
            .finally(() => setSaving(false));
    };

    const handleUpdate = (updates: Partial<VideoSettingsType>) => {
        if (!settings) return;
        const newSettings = { ...settings, ...updates };
        setSettings(newSettings);
        saveSettings(newSettings);
    };

    if (loading) return <Loader />;

    // Fallback safe values
    const safeResolution = settings?.resolution || '';
    const safeFps = settings?.fps || 0;
    const safeBitrate = settings?.bitrate || 0;
    const safeGopSize = settings?.gop_size || 0;
    const safeCodec = settings?.codec || '';

    const isDisabled = saving || connectionError;

    return (
        <Stack gap="md">
            <div>
                <Text size="sm" fw={500} mb={3}>Resolution</Text>
                <Select
                    value={safeResolution}
                    onChange={(val) => {
                        if (val) handleUpdate({ resolution: val });
                    }}
                    data={[
                        { value: '1920x1080', label: '1080p (1920x1080)' },
                        { value: '1280x720', label: '720p (1280x720)' },
                    ]}
                    disabled={isDisabled}
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>FPS</Text>
                <Select
                    value={safeFps ? safeFps.toString() : ''}
                    onChange={(val) => handleUpdate({ fps: Number(val) })}
                    data={['60', '90', '120']}
                    disabled={isDisabled}
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>Bitrate (kbps)</Text>
                <NumberInput
                    value={safeBitrate}
                    onChange={(val) => setSettings(prev => prev ? ({ ...prev, bitrate: Number(val) }) : null)}
                    onBlur={() => settings && saveSettings(settings)}
                    step={500}
                    disabled={isDisabled}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>GOP Size</Text>
                <NumberInput
                    value={safeGopSize}
                    onChange={(val) => setSettings(prev => prev ? ({ ...prev, gop_size: Number(val) }) : null)}
                    onBlur={() => settings && saveSettings(settings)}
                    min={1}
                    disabled={isDisabled}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>Codec</Text>
                <Select
                    value={safeCodec}
                    onChange={(val) => handleUpdate({ codec: val as 'h264' | 'h265' })}
                    data={['h264', 'h265']}
                    disabled={isDisabled}
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>
        </Stack>
    );
}
