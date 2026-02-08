import { useState, useEffect } from 'react';
import { Stack, Switch, Loader, Text, Select } from '@mantine/core';
import type { AdaptiveLinkSettings } from '../types';
import { fetchWithTimeout } from '../utils/api';

interface SystemSettingsProps {
    onAlinkChange?: (enabled: boolean) => void;
}

export function SystemSettings({ onAlinkChange }: SystemSettingsProps) {
    const [alinkSettings, setAlinkSettings] = useState<AdaptiveLinkSettings | null>(null);
    const [connectionError, setConnectionError] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const alinkRes = await fetchWithTimeout('/api/v1/adaptive-link');
                if (!alinkRes.ok) throw new Error('Network response was not ok');
                const alinkData = await alinkRes.json();

                setAlinkSettings(alinkData);
                if (onAlinkChange) onAlinkChange(alinkData.enabled);
                setConnectionError(false);
            } catch (err) {
                console.warn('Failed to fetch alink settings', err);
                setConnectionError(true);
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const saveAlinkSettings = (newSettings: AdaptiveLinkSettings) => {
        setSaving(true);
        fetchWithTimeout('/api/v1/adaptive-link', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newSettings),
        })
            .then(() => {
                if (onAlinkChange) onAlinkChange(newSettings.enabled);
            })
            .finally(() => setSaving(false));
    };

    const handleUpdate = (updates: Partial<AdaptiveLinkSettings>) => {
        if (!alinkSettings) return;
        const newSettings = { ...alinkSettings, ...updates };
        setAlinkSettings(newSettings);
        saveAlinkSettings(newSettings);
    };

    if (loading) return <Loader />;

    // Fallback safe values
    const safeEnabled = alinkSettings?.enabled || false;
    const safeAllowSetPower = alinkSettings?.allow_set_power || false;
    const safePowerLevel = alinkSettings?.power_level_0_to_4 || 0;

    const isDisabled = saving || connectionError;

    return (
        <Stack gap="xl">
            <Stack gap="md">
                <Text size="lg" fw={700}>Adaptive Link</Text>
                <Switch
                    label="Enable Adaptive Link"
                    checked={safeEnabled}
                    onChange={(event) => handleUpdate({ enabled: event.currentTarget.checked })}
                    disabled={isDisabled}
                />
                <Switch
                    label="Allow Set Power"
                    checked={safeAllowSetPower}
                    onChange={(event) => handleUpdate({ allow_set_power: event.currentTarget.checked })}
                    disabled={!safeEnabled || isDisabled}
                />
                <div>
                    <Text size="sm" fw={500} mb={3}>Power Level</Text>
                    <Select
                        value={safePowerLevel.toString()}
                        onChange={(val) => handleUpdate({ power_level_0_to_4: Number(val) })}
                        data={['0', '1', '2', '3', '4']}
                        disabled={!safeEnabled || isDisabled}
                        comboboxProps={{ zIndex: 2100 }}
                    />
                </div>
            </Stack>
            <div>
                <Text size="sm" fw={500} mb={3}>OSD Level</Text>
                <Select
                    value={(alinkSettings?.osd_level ?? 1).toString()}
                    onChange={(val) => handleUpdate({ osd_level: Number(val) })}
                    data={['1', '2', '3', '4', '5', '6']}
                    disabled={!safeEnabled || isDisabled}
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>
        </Stack>
    );
}
