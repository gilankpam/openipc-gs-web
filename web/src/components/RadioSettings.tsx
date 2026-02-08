import { useState, useEffect } from 'react';
import { Stack, Slider, Text, NumberInput, Select, Loader, Group, Alert } from '@mantine/core';
import type { RadioSettings as RadioSettingsType } from '../types';
import { fetchWithTimeout } from '../utils/api';

interface RadioSettingsProps {
    alinkEnabled: boolean;
}

const WIFI_5GHZ_CHANNELS = [
    { channel: 36, freq: 5180 },
    { channel: 40, freq: 5200 },
    { channel: 44, freq: 5220 },
    { channel: 48, freq: 5240 },
    { channel: 52, freq: 5260 },
    { channel: 56, freq: 5280 },
    { channel: 60, freq: 5300 },
    { channel: 64, freq: 5320 },
    { channel: 100, freq: 5500 },
    { channel: 104, freq: 5520 },
    { channel: 108, freq: 5540 },
    { channel: 112, freq: 5560 },
    { channel: 116, freq: 5580 },
    { channel: 120, freq: 5600 },
    { channel: 124, freq: 5620 },
    { channel: 128, freq: 5640 },
    { channel: 132, freq: 5660 },
    { channel: 136, freq: 5680 },
    { channel: 140, freq: 5700 },
    { channel: 144, freq: 5720 },
    { channel: 149, freq: 5745 },
    { channel: 153, freq: 5765 },
    { channel: 157, freq: 5785 },
    { channel: 161, freq: 5805 },
    { channel: 165, freq: 5825 },
];

export function RadioSettings({ alinkEnabled }: RadioSettingsProps) {
    const [settings, setSettings] = useState<RadioSettingsType | null>(null);
    const [connectionError, setConnectionError] = useState(false);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [isLocalOnly, setIsLocalOnly] = useState(false);

    useEffect(() => {
        const fetchSettings = async () => {
            try {
                const radioRes = await fetchWithTimeout('/api/v1/radio', { timeout: 5000 });
                if (!radioRes.ok) throw new Error('Network response was not ok');

                const dataSource = radioRes.headers.get('X-GS-Data-Source');
                setIsLocalOnly(dataSource === 'local');

                const radioData = await radioRes.json();
                setSettings(radioData);
                setConnectionError(false);
            } catch (err) {
                console.warn('Failed to fetch settings', err);
                setConnectionError(true);
            } finally {
                setLoading(false);
            }
        };
        fetchSettings();
    }, []);

    const saveSettings = (newSettings: RadioSettingsType) => {
        setSaving(true);
        fetchWithTimeout('/api/v1/radio', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(newSettings),
            timeout: 5000,
        })
            .then((res) => {
                if (res.ok) {
                    const dataSource = res.headers.get('X-GS-Data-Source');
                    setIsLocalOnly(dataSource === 'local');
                }
            })
            .finally(() => setSaving(false));
    };

    const handleUpdate = (key: keyof RadioSettingsType, value: any) => {
        if (!settings) return;
        const newSettings = { ...settings, [key]: value };
        setSettings(newSettings);
        saveSettings(newSettings);
    };

    if (loading) return <Loader />;

    // Fallback safe values
    const safeChannel = settings?.channel || 0;
    const safeBandwidth = settings?.bandwidth || 0;
    const safeTxPower = settings?.tx_power || 0;
    const safeMcsIndex = settings?.mcs_index || 0;
    const safeFecK = settings?.fec_k || 0;
    const safeFecN = settings?.fec_n || 0;

    const isDisabled = saving || connectionError;
    // If local only, we enable channel but disable others?
    // Actually, if isLocalOnly is true, connectionError should be false (we handle it).
    // So 'isDisabled' will be false (unless saving).
    // We want to specifically disable non-channel inputs if isLocalOnly is true.

    return (
        <Stack gap="md">
            {isLocalOnly && (
                <Alert variant="light" color="orange" title="Air Unit Not Connected">
                    Settings are currently being saved to the Ground Station only.
                    Only Channel changes are allowed and will take effect after a local service restart.
                </Alert>
            )}

            <div>
                <Text size="sm" fw={500} mb={3}>Channel</Text>
                <Select
                    value={safeChannel ? safeChannel.toString() : ''}
                    onChange={(val) => handleUpdate('channel', Number(val))}
                    data={[
                        ...WIFI_5GHZ_CHANNELS.map(c => ({
                            value: c.channel.toString(),
                            label: `Channel ${c.channel} (${c.freq} MHz)`
                        })),
                        ...(safeChannel && !WIFI_5GHZ_CHANNELS.find(c => c.channel === safeChannel)
                            ? [{ value: safeChannel.toString(), label: `Custom (${safeChannel})` }]
                            : [])
                    ]}
                    searchable
                    disabled={isDisabled} // Always allow unless saving or fully errored
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>

            <div>
                <Text size="sm" fw={500} mb={3}>Bandwidth</Text>
                <Select
                    value={safeBandwidth ? safeBandwidth.toString() : ''}
                    onChange={(val) => handleUpdate('bandwidth', Number(val))}
                    data={['20', '40', '80']}
                    disabled={isDisabled || isLocalOnly}
                    comboboxProps={{ zIndex: 2100 }}
                />
            </div>

            {!alinkEnabled && (
                <>
                    <div>
                        <Text size="sm" fw={500} mb={3}>TX Power</Text>
                        <Slider
                            value={safeTxPower}
                            onChange={(val) => setSettings(prev => prev ? ({ ...prev, tx_power: val }) : null)}
                            onChangeEnd={(val) => settings && saveSettings({ ...settings, tx_power: val })}
                            min={1}
                            max={60}
                            marks={[
                                { value: 10, label: '10' },
                                { value: 30, label: '30' },
                                { value: 50, label: '50' },
                            ]}
                            disabled={isDisabled || isLocalOnly}
                        />
                    </div>

                    <Group grow>
                        <div>
                            <Text size="sm" fw={500} mb={3}>MCS Index</Text>
                            <NumberInput
                                value={safeMcsIndex}
                                onChange={(val) => setSettings(prev => prev ? ({ ...prev, mcs_index: Number(val) }) : null)}
                                onBlur={() => settings && saveSettings(settings)}
                                disabled={isDisabled || isLocalOnly}
                            />
                        </div>
                        <div>
                            <Text size="sm" fw={500} mb={3}>FEC K</Text>
                            <NumberInput
                                value={safeFecK}
                                onChange={(val) => setSettings(prev => prev ? ({ ...prev, fec_k: Number(val) }) : null)}
                                onBlur={() => settings && saveSettings(settings)}
                                disabled={isDisabled || isLocalOnly}
                            />
                        </div>
                        <div>
                            <Text size="sm" fw={500} mb={3}>FEC N</Text>
                            <NumberInput
                                value={safeFecN}
                                onChange={(val) => setSettings(prev => prev ? ({ ...prev, fec_n: Number(val) }) : null)}
                                onBlur={() => settings && saveSettings(settings)}
                                disabled={isDisabled || isLocalOnly}
                            />
                        </div>
                    </Group>
                </>
            )}
        </Stack>
    );
}
