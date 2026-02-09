import { ActionIcon, Box, Group, Paper, Text, Tooltip } from '@mantine/core';
import { IconAccessPoint, IconAntenna, IconChartLine, IconEye, IconRefresh } from '@tabler/icons-react';
import { useEffect, useRef, useState } from 'react';
// @ts-ignore
import Draggable from 'react-draggable';
import type { WFBStats as WFBStatsType } from '../types';

export function WFBStats() {
    const [stats, setStats] = useState<WFBStatsType | null>(null);
    const [visible, setVisible] = useState(true);
    const nodeRef = useRef(null);

    useEffect(() => {
        const interval = setInterval(async () => {
            if (!visible) return;
            try {
                const res = await fetch('/api/v1/stats');
                if (res.ok) {
                    const data = await res.json();
                    setStats(data);
                }
            } catch (err) {
                console.error("Failed to fetch stats", err);
            }
        }, 1000);

        return () => clearInterval(interval);
    }, [visible]);

    if (!visible) {
        return (
            <Box style={{ position: 'absolute', bottom: 20, left: 20, zIndex: 100 }}>
                <Tooltip label="Show Stats">
                    <ActionIcon onClick={() => setVisible(true)} variant="filled" color="dark" size="lg" radius="xl" style={{ opacity: 0.8 }}>
                        <IconChartLine size={20} />
                    </ActionIcon>
                </Tooltip>
            </Box>
        );
    }

    return (
        <Draggable nodeRef={nodeRef}>
            <Box ref={nodeRef} style={{ position: 'absolute', bottom: 20, left: 20, zIndex: 1000, cursor: 'move', touchAction: 'none' }}>
                <Paper p="xs" radius="md" style={{ backgroundColor: 'rgba(0, 0, 0, 0.6)', color: 'white', backdropFilter: 'blur(4px)', minWidth: 200 }}>
                    <Group justify="space-between" mb={5}>
                        <Group gap={5}>
                            <IconAccessPoint size={16} />
                            <Text size="xs" fw={700}>WFB-NG Stats</Text>
                        </Group>
                        <ActionIcon onClick={() => setVisible(false)} variant="subtle" color="gray" size="xs">
                            <IconEye size={14} />
                        </ActionIcon>
                    </Group>

                    {stats ? (
                        <>
                            {/* Antenna Stats */}
                            <Box mb={8}>
                                <Group gap={5} mb={2}>
                                    <IconAntenna size={14} style={{ opacity: 0.7 }} />
                                    <Text size="xs" c="dimmed">Signal (RSSI / SNR)</Text>
                                </Group>
                                <Group gap={8}>
                                    {stats.rssi && stats.rssi.map((rssi, i) => (
                                        <Text key={`rssi-${i}`} size="xs" span>
                                            A{i + 1}: <Text span c={rssi > -60 ? "lime" : rssi > -75 ? "yellow" : "red"} fw={700}>{rssi}</Text> / <Text span c="cyan">{stats.snr[i]}</Text>
                                        </Text>
                                    ))}
                                    {(!stats.rssi || stats.rssi.length === 0) && <Text size="xs" c="dimmed">No signal data</Text>}
                                </Group>
                            </Box>

                            {/* Packet Stats */}
                            <Box mb={8}>
                                <Group gap={5} mb={2}>
                                    <IconChartLine size={14} style={{ opacity: 0.7 }} />
                                    <Text size="xs" c="dimmed">Packets / Sec</Text>
                                </Group>
                                <Group gap={15} align="center">
                                    <Box>
                                        <Text size="xs" c="dimmed" lh={1}>Video</Text>
                                        <Text size="sm" fw={700}>{stats.video_packets_per_sec}</Text>
                                    </Box>
                                    <Box>
                                        <Text size="xs" c="dimmed" lh={1}>FEC</Text>
                                        <Text size="sm" fw={700} c="cyan">{stats.fec_packets_per_sec}</Text>
                                    </Box>
                                    <Box>
                                        <Text size="xs" c="dimmed" lh={1}>Lost</Text>
                                        <Text size="sm" fw={700} c={stats.lost_packets_per_sec > 0 ? "red" : "gray"}>{stats.lost_packets_per_sec}</Text>
                                    </Box>
                                </Group>
                            </Box>

                            {/* Transmission Info */}
                            <Box style={{ borderTop: '1px solid rgba(255, 255, 255, 0.1)', paddingTop: 4 }}>
                                <Group justify="space-between" mb={2}>
                                    <Text size="xs" c="dimmed">Radio</Text>
                                    <Text size="xs" fw={500} c="white">
                                        {stats.frequency}MHz / {stats.bandwidth}MHz / MCS {stats.mcs_index}
                                    </Text>
                                </Group>
                                <Group justify="space-between" mb={2}>
                                    <Text size="xs" c="dimmed">FEC Ratio</Text>
                                    <Text size="xs" fw={500} c="white">
                                        {stats.fec_k} / {stats.fec_n}
                                    </Text>
                                </Group>
                                <Group justify="space-between">
                                    <Text size="xs" c="dimmed">Flow</Text>
                                    <Text size="xs" fw={500} c="gold">
                                        {(stats.link_flow_bytes_per_sec * 8 / 1000 / 1000).toFixed(1)} Mbit/s
                                    </Text>
                                </Group>
                            </Box>
                        </>
                    ) : (
                        <Group justify="center" p="xs">
                            <IconRefresh size={16} className="mantine-rotate" />
                            <Text size="xs">Waiting for data...</Text>
                        </Group>
                    )}
                </Paper>
            </Box>
        </Draggable>
    );
}
