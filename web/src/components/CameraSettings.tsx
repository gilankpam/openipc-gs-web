import { useState, useEffect } from 'react';
import { Stack, Switch, Group, Loader, Text, Slider } from '@mantine/core';
import type { CameraSettings as CameraSettingsType } from '../types';

export function CameraSettings() {
    const [settings, setSettings] = useState<CameraSettingsType | null>(null);
    const [connectionError, setConnectionError] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const cameraRes = await fetch('/api/v1/camera');
                if (!cameraRes.ok) throw new Error('Network response was not ok');
                const cameraData = await cameraRes.json();

                setSettings(cameraData);
                setConnectionError(false);
            } catch (err) {
                console.warn('Failed to fetch camera settings', err);
                setConnectionError(true);
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const saveSettings = (newSettings: CameraSettingsType) => {
        setSaving(true);
        fetch('/api/v1/camera', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newSettings),
        })
            .finally(() => setSaving(false));
    };

    const handleUpdate = (updates: Partial<CameraSettingsType>) => {
        if (!settings) return;
        const newSettings = { ...settings, ...updates };
        setSettings(newSettings);
        saveSettings(newSettings);
    };

    if (loading) return <Loader />;

    // Fallback safe values
    const safeContrast = settings?.contrast || 50;
    const safeSaturation = settings?.saturation || 50;
    const safeFlip = settings?.flip || false;
    const safeMirror = settings?.mirror || false;

    const isDisabled = saving || connectionError;

    return (
        <Stack gap="md">
            <div>
                <Text size="sm" fw={500} mb={3}>Contrast</Text>
                <Slider
                    value={safeContrast}
                    onChange={(val) => setSettings(prev => prev ? ({ ...prev, contrast: val }) : null)}
                    onChangeEnd={(val) => settings && saveSettings({ ...settings, contrast: val })}
                    disabled={isDisabled}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>Saturation</Text>
                <Slider
                    value={safeSaturation}
                    onChange={(val) => setSettings(prev => prev ? ({ ...prev, saturation: val }) : null)}
                    onChangeEnd={(val) => settings && saveSettings({ ...settings, saturation: val })}
                    disabled={isDisabled}
                />
            </div>

            <Group mt="xs">
                <Switch
                    label="Flip Image"
                    checked={safeFlip}
                    onChange={(event) => handleUpdate({ flip: event.currentTarget.checked })}
                    disabled={isDisabled}
                />
                <Switch
                    label="Mirror Image"
                    checked={safeMirror}
                    onChange={(event) => handleUpdate({ mirror: event.currentTarget.checked })}
                    disabled={isDisabled}
                />
            </Group>
        </Stack>
    );
}
