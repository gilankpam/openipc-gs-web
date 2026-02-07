import { useState, useEffect, useRef } from 'react';
import { Stack, Paper, Text, Group, Button, Grid, NumberInput, Select, TextInput, Divider, Loader, Box, Tooltip } from '@mantine/core';
import { IconDeviceFloppy, IconPlus, IconTrash, IconRotateClockwise } from '@tabler/icons-react';
import type { TxProfile } from '../types';

const MIN_RANGE = 999;
const MAX_RANGE = 2000;
const STEP = 50;

const DEFAULT_PROFILES: TxProfile[] = [
    { range_start: 999, range_end: 999, gi: 'long', mcs: 0, fec_k: 2, fec_n: 3, bitrate: 1000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
    { range_start: 1000, range_end: 1050, gi: 'long', mcs: 0, fec_k: 2, fec_n: 3, bitrate: 2000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
    { range_start: 1051, range_end: 1100, gi: 'long', mcs: 1, fec_k: 2, fec_n: 3, bitrate: 4000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
    { range_start: 1101, range_end: 1200, gi: 'long', mcs: 2, fec_k: 4, fec_n: 6, bitrate: 7000, gop: 10, pwr: 30, roi_qp: '12,8,8,12', bandwidth: 20, qp_delta: -12 },
    { range_start: 1201, range_end: 1300, gi: 'long', mcs: 3, fec_k: 6, fec_n: 9, bitrate: 10000, gop: 10, pwr: 30, roi_qp: '2,1,1,2', bandwidth: 20, qp_delta: -12 },
    { range_start: 1301, range_end: 1400, gi: 'long', mcs: 4, fec_k: 6, fec_n: 9, bitrate: 13000, gop: 10, pwr: 30, roi_qp: '2,1,1,2', bandwidth: 20, qp_delta: -12 },
    { range_start: 1401, range_end: 1600, gi: 'short', mcs: 4, fec_k: 8, fec_n: 12, bitrate: 14000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
    { range_start: 1601, range_end: 1800, gi: 'long', mcs: 4, fec_k: 10, fec_n: 15, bitrate: 15000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
    { range_start: 1801, range_end: 2001, gi: 'short', mcs: 4, fec_k: 11, fec_n: 15, bitrate: 19000, gop: 10, pwr: 30, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12 },
];

export function TxProfilesSettings() {
    const [profiles, setProfiles] = useState<TxProfile[]>([]);
    const [loading, setLoading] = useState(true);
    const [saving, setSaving] = useState(false);
    const [connectionError, setConnectionError] = useState(false);
    const [isDirty, setIsDirty] = useState(false);
    const [selectedIndex, setSelectedIndex] = useState<number | null>(null);

    // draggingIndex tracks which boundary (handle) is being dragged.
    // Index i corresponds to the boundary between profiles[i] and profiles[i+1].
    // so draggingIndex 0 is the boundary after the first profile.
    const [draggingIndex, setDraggingIndex] = useState<number | null>(null);
    const trackRef = useRef<HTMLDivElement>(null);

    useEffect(() => {
        fetchProfiles();
    }, []);

    const fetchProfiles = async () => {
        try {
            const res = await fetch('/api/v1/txprofiles');
            if (!res.ok) throw new Error('Failed to fetch profiles');
            let data: TxProfile[] = await res.json();

            if (!Array.isArray(data) || data.length === 0) {
                data = [{
                    range_start: MIN_RANGE,
                    range_end: MAX_RANGE,
                    gi: 'long', mcs: 0, fec_k: 8, fec_n: 12, bitrate: 5000, gop: 2, pwr: 20, roi_qp: '0,0,0,0', bandwidth: 20, qp_delta: -12
                }];
            } else {
                data.sort((a, b) => a.range_start - b.range_start);
                // Ensure continuity and bounds
                if (data[0].range_start !== MIN_RANGE) data[0].range_start = MIN_RANGE;
                if (data[data.length - 1].range_end !== MAX_RANGE) data[data.length - 1].range_end = MAX_RANGE;

                // Fix gaps: Next.Start = Prev.End + 1
                for (let i = 0; i < data.length - 1; i++) {
                    data[i + 1].range_start = data[i].range_end + 1;
                }
            }

            setProfiles(data);
            setSelectedIndex(0);
            setConnectionError(false);
            setIsDirty(false);
        } catch (err) {
            console.warn('Error fetching txprofiles:', err);
            setConnectionError(true);
        } finally {
            setLoading(false);
        }
    };

    const handleSave = async () => {
        setSaving(true);
        try {
            const res = await fetch('/api/v1/txprofiles', {
                method: 'POST',
                headers: { 'Content-Type': 'application/json' },
                body: JSON.stringify(profiles),
            });
            if (!res.ok) throw new Error('Failed to save profiles');
            setIsDirty(false);
        } catch (err) {
            console.error('Error saving txprofiles:', err);
        } finally {
            setSaving(false);
        }
    };

    const handleReset = () => {
        if (!window.confirm("Are you sure you want to reset all profiles to default settings? (You must save changes to apply)")) return;
        setProfiles([...DEFAULT_PROFILES]);
        setSelectedIndex(0);
        setIsDirty(true);
    };

    const updateProfile = (field: keyof TxProfile, value: any) => {
        if (selectedIndex === null) return;
        const updated = [...profiles];
        updated[selectedIndex] = { ...updated[selectedIndex], [field]: value };
        setProfiles(updated);
        setIsDirty(true);
    };

    // Split the current selected profile
    const handleSplit = () => {
        if (selectedIndex === null) return;
        const current = profiles[selectedIndex];
        const span = current.range_end - current.range_start;
        // Need space for at least 2 ranges of size STEP? Or just minimal valid split
        if (span < STEP * 2) return;

        // Split in half, snapped to STEP
        let mid = current.range_start + Math.floor((span / 2) / STEP) * STEP;

        // Constraints
        if (mid <= current.range_start) mid = current.range_start + STEP;
        if (mid >= current.range_end) mid = current.range_end - STEP;

        const firstHalf = { ...current, range_end: mid };
        const secondHalf = { ...current, range_start: mid + 1 }; // Next starts at +1

        const updated = [
            ...profiles.slice(0, selectedIndex),
            firstHalf,
            secondHalf,
            ...profiles.slice(selectedIndex + 1)
        ];

        setProfiles(updated);
        setSelectedIndex(selectedIndex + 1);
        setIsDirty(true);
    };

    const handleMerge = () => {
        if (selectedIndex === null || profiles.length <= 1) return;

        const updated = [...profiles];
        if (selectedIndex > 0) {
            // Merge with previous
            const prev = updated[selectedIndex - 1];
            const curr = updated[selectedIndex];
            prev.range_end = curr.range_end;
            updated.splice(selectedIndex, 1);
            setProfiles(updated);
            setSelectedIndex(selectedIndex - 1);
        } else {
            // Merge with next (idx 0 case)
            const curr = updated[0];
            const next = updated[1];
            curr.range_end = next.range_end;
            updated.splice(1, 1);
            setProfiles(updated);
        }
        setIsDirty(true);
    };

    // --- Drag Logic ---

    const valueToPercent = (val: number) => {
        return ((val - MIN_RANGE) / (MAX_RANGE - MIN_RANGE)) * 100;
    };

    const handlePointerDown = (e: React.PointerEvent, index: number) => {
        e.preventDefault();
        e.stopPropagation(); // Prevent clicking the track underneath
        setDraggingIndex(index);
        (e.target as HTMLElement).setPointerCapture(e.pointerId);
    };

    const handlePointerMove = (e: React.PointerEvent) => {
        if (draggingIndex === null || !trackRef.current) return;

        const rect = trackRef.current.getBoundingClientRect();
        const x = e.clientX - rect.left;
        const width = rect.width;
        let p = x / width;
        // Clamp 0..1
        if (p < 0) p = 0;
        if (p > 1) p = 1;

        let rawVal = MIN_RANGE + p * (MAX_RANGE - MIN_RANGE);
        // Snap to step
        let newVal = Math.round(rawVal / STEP) * STEP;

        // Constraints
        // profiles[draggingIndex] is Left Profile. We are setting its End.
        // profiles[draggingIndex+1] is Right Profile. We are setting its Start to End + 1.

        const leftProfile = profiles[draggingIndex];
        const rightProfile = profiles[draggingIndex + 1];

        // left.End must be >= left.Start
        // right.Start = left.End + 1
        // right.Start must be <= right.End  => left.End + 1 <= right.End => left.End <= right.End - 1

        const lowerBound = leftProfile.range_start + STEP;
        const upperBound = rightProfile.range_end - STEP;

        if (newVal < lowerBound) newVal = lowerBound;
        if (newVal > upperBound) newVal = upperBound;

        if (newVal !== leftProfile.range_end) {
            const updated = [...profiles];
            updated[draggingIndex].range_end = newVal;
            updated[draggingIndex + 1].range_start = newVal + 1;
            setProfiles(updated);
            setIsDirty(true);
        }
    };

    const handlePointerUp = (e: React.PointerEvent) => {
        if (draggingIndex !== null) {
            setDraggingIndex(null);
            (e.target as HTMLElement).releasePointerCapture(e.pointerId);
        }
    };

    if (loading) return <Loader />;
    if (connectionError) return <Text c="dimmed" fs="italic">Tx Profiles settings unavailable (Disconnected)</Text>;

    const isDisabled = saving || connectionError;
    const selectedProfile = selectedIndex !== null ? profiles[selectedIndex] : null;

    return (
        <Stack gap="xl">
            {/* Custom Multi-Handle Slider */}
            <Box pt="lg" pb="xs">
                <Text size="sm" fw={500} mb="sm">Profile Ranges (Drag red handles to adjust)</Text>
                <div
                    ref={trackRef}
                    style={{
                        position: 'relative',
                        height: '40px',
                        width: '100%',
                        marginTop: '20px',
                        marginBottom: '10px'
                    }}
                >
                    {/* Track Background Line */}
                    <div style={{
                        position: 'absolute',
                        top: '50%',
                        left: '0',
                        right: '0',
                        height: '4px',
                        backgroundColor: '#495057', // dark-4
                        transform: 'translateY(-50%)',
                        borderRadius: '2px'
                    }} />

                    {/* Segments (Clickable Areas) */}
                    {profiles.map((p, i) => {
                        const startPct = valueToPercent(p.range_start);
                        const endPct = valueToPercent(p.range_end);
                        const widthPct = endPct - startPct;
                        const isSelected = i === selectedIndex;

                        return (
                            <Tooltip key={`seg-${i}`} label={`${p.range_start} - ${p.range_end}`} position="top" opened={isSelected}>
                                <div
                                    onClick={() => setSelectedIndex(i)}
                                    style={{
                                        position: 'absolute',
                                        left: `${startPct}%`,
                                        width: `${widthPct}%`,
                                        top: '50%',
                                        height: '12px', // Thicker than track line to be clickable
                                        transform: 'translateY(-50%)',
                                        backgroundColor: isSelected ? 'var(--mantine-color-blue-filled)' : 'transparent',
                                        cursor: 'pointer',
                                        zIndex: 1,
                                        borderRadius: '4px',
                                        border: isSelected ? '1px solid white' : 'none'
                                    }}
                                />
                            </Tooltip>
                        );
                    })}

                    {/* Handles (Draggable) */}
                    {profiles.slice(0, -1).map((p, i) => {
                        // Render handle at p.range_end
                        const posPct = valueToPercent(p.range_end);
                        return (
                            <div
                                key={`handle-${i}`}
                                onPointerDown={(e) => handlePointerDown(e, i)}
                                onPointerMove={handlePointerMove}
                                onPointerUp={handlePointerUp}
                                style={{
                                    position: 'absolute',
                                    left: `calc(${posPct}% - 12px)`, // Center the 24px handle
                                    top: '50%',
                                    marginTop: '-12px',
                                    width: '24px',
                                    height: '24px',
                                    borderRadius: '50%',
                                    backgroundColor: '#fa5252', // red-6
                                    border: '2px solid white',
                                    cursor: 'ew-resize',
                                    zIndex: 10,
                                    display: 'flex',
                                    alignItems: 'center',
                                    justifyContent: 'center',
                                    boxShadow: '0 2px 4px rgba(0,0,0,0.3)'
                                }}
                            >
                                <Text size="0.6rem" c="white" fw={700}>||</Text>
                                {/* Floating Label above handle */}
                                <div style={{
                                    position: 'absolute',
                                    top: '-25px',
                                    left: '50%',
                                    transform: 'translateX(-50%)',
                                    whiteSpace: 'nowrap',
                                    fontSize: '10px',
                                    color: '#fa5252',
                                    fontWeight: 'bold',
                                    pointerEvents: 'none'
                                }}>
                                    {p.range_end}
                                </div>
                            </div>
                        );
                    })}

                    {/* Fixed Start/End Labels */}
                    <Text
                        size="xs" c="dimmed"
                        style={{ position: 'absolute', top: '100%', left: 0, marginTop: '4px' }}
                    >
                        {MIN_RANGE}
                    </Text>
                    <Text
                        size="xs" c="dimmed"
                        style={{ position: 'absolute', top: '100%', right: 0, marginTop: '4px' }}
                    >
                        {MAX_RANGE}
                    </Text>
                </div>
            </Box>

            {selectedProfile && (
                <Paper p="md" withBorder>
                    <Stack gap="xs">
                        <Group justify="space-between">
                            <Text fw={700}>Selected Segment: {selectedProfile.range_start} - {selectedProfile.range_end}</Text>
                            <Group gap="xs">
                                <Button
                                    size="xs"
                                    variant="light"
                                    onClick={handleSplit}
                                    disabled={isDisabled || (selectedProfile.range_end - selectedProfile.range_start) < STEP * 2}
                                    leftSection={<IconPlus size={14} />}
                                >
                                    Split Range
                                </Button>
                                <Button
                                    size="xs"
                                    variant="subtle"
                                    color="red"
                                    onClick={handleMerge}
                                    disabled={isDisabled || profiles.length <= 1}
                                    leftSection={<IconTrash size={14} />}
                                >
                                    Merge
                                </Button>
                            </Group>
                        </Group>

                        <Divider my="xs" label="Transmission Parameters" labelPosition="center" />

                        <Grid>
                            <Grid.Col span={4}>
                                <Select
                                    label="GI"
                                    data={['long', 'short']}
                                    value={selectedProfile.gi}
                                    onChange={(v) => updateProfile('gi', v)}
                                    disabled={isDisabled}
                                    comboboxProps={{ zIndex: 2100 }}
                                />
                            </Grid.Col>
                            <Grid.Col span={4}>
                                <Select
                                    label="MCS"
                                    data={Array.from({ length: 8 }, (_, i) => i.toString())}
                                    value={selectedProfile.mcs.toString()}
                                    onChange={(v) => updateProfile('mcs', Number(v))}
                                    disabled={isDisabled}
                                    comboboxProps={{ zIndex: 2100 }}
                                />
                            </Grid.Col>
                            <Grid.Col span={4}>
                                <Select
                                    label="Bandwidth"
                                    data={['10', '20', '40']}
                                    value={selectedProfile.bandwidth.toString()}
                                    onChange={(v) => updateProfile('bandwidth', Number(v))}
                                    disabled={isDisabled}
                                    comboboxProps={{ zIndex: 2100 }}
                                />
                            </Grid.Col>

                            <Grid.Col span={3}>
                                <NumberInput label="FEC K" value={selectedProfile.fec_k} onChange={(v) => updateProfile('fec_k', Number(v))} disabled={isDisabled} hideControls />
                            </Grid.Col>
                            <Grid.Col span={3}>
                                <NumberInput label="FEC N" value={selectedProfile.fec_n} onChange={(v) => updateProfile('fec_n', Number(v))} disabled={isDisabled} hideControls />
                            </Grid.Col>
                            <Grid.Col span={6}>
                                <NumberInput label="Bitrate" value={selectedProfile.bitrate} onChange={(v) => updateProfile('bitrate', Number(v))} disabled={isDisabled} />
                            </Grid.Col>

                            <Grid.Col span={3}>
                                <NumberInput label="GOP" value={selectedProfile.gop} onChange={(v) => updateProfile('gop', Number(v))} disabled={isDisabled} hideControls />
                            </Grid.Col>
                            <Grid.Col span={3}>
                                <NumberInput label="Power" value={selectedProfile.pwr} onChange={(v) => updateProfile('pwr', Number(v))} disabled={isDisabled} hideControls />
                            </Grid.Col>
                            <Grid.Col span={3}>
                                <NumberInput label="QP Delta" value={selectedProfile.qp_delta} onChange={(v) => updateProfile('qp_delta', Number(v))} disabled={isDisabled} hideControls />
                            </Grid.Col>
                            <Grid.Col span={3}>
                                <TextInput label="ROI QP" value={selectedProfile.roi_qp} onChange={(e) => updateProfile('roi_qp', e.currentTarget.value)} disabled={isDisabled} />
                            </Grid.Col>
                        </Grid>
                    </Stack>
                </Paper>
            )}

            <Group mt="md" grow>
                <Button
                    variant="light"
                    color="orange"
                    leftSection={<IconRotateClockwise size={16} />}
                    onClick={handleReset}
                    disabled={saving || connectionError}
                >
                    Reset to Defaults
                </Button>
                <Button
                    leftSection={<IconDeviceFloppy size={16} />}
                    onClick={handleSave}
                    loading={saving}
                    disabled={isDisabled || !isDirty}
                >
                    Save All Changes
                </Button>
            </Group>
        </Stack>
    );
}
