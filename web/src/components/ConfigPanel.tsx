import { useState, useEffect } from 'react';
import { Tabs, rem, Paper } from '@mantine/core';
import { IconRadio, IconVideo, IconCamera, IconActivity } from '@tabler/icons-react';
import { RadioSettings } from './RadioSettings';
import { VideoSettings } from './VideoSettings';
import { CameraSettings } from './CameraSettings';
import { SystemSettings } from './SystemSettings';
import { TxProfilesSettings } from './TxProfilesSettings';
import { fetchWithTimeout } from '../utils/api';

interface ConfigPanelProps {
    isConnected: boolean;
}

export function ConfigPanel({ isConnected }: ConfigPanelProps) {
    const iconStyle = { width: rem(12), height: rem(12) };
    const [alinkEnabled, setAlinkEnabled] = useState(false);

    useEffect(() => {
        if (isConnected) {
            fetchWithTimeout('/api/v1/adaptive-link')
                .then(res => res.json())
                .then(data => setAlinkEnabled(data.enabled))
                .catch(console.error);
        }
    }, [isConnected]);

    return (
        <Paper shadow="sm" radius="md" p="md" withBorder>
            <Tabs defaultValue="radio" variant="outline" key={isConnected ? 'connected' : 'disconnected'}>
                <Tabs.List>
                    <Tabs.Tab value="radio" leftSection={<IconRadio style={iconStyle} />}>
                        Radio
                    </Tabs.Tab>
                    <Tabs.Tab value="video" leftSection={<IconVideo style={iconStyle} />}>
                        Video
                    </Tabs.Tab>
                    <Tabs.Tab value="camera" leftSection={<IconCamera style={iconStyle} />}>
                        Camera
                    </Tabs.Tab>
                    <Tabs.Tab value="alink" leftSection={<IconActivity style={iconStyle} />}>
                        Adaptive Link
                    </Tabs.Tab>
                    <Tabs.Tab
                        value="txprofiles"
                        leftSection={<IconActivity style={iconStyle} />}
                        disabled={!alinkEnabled}
                    >
                        Tx Profiles
                    </Tabs.Tab>
                </Tabs.List>

                <Tabs.Panel value="radio" pt="xs">
                    <RadioSettings alinkEnabled={alinkEnabled} />
                </Tabs.Panel>
                <Tabs.Panel value="video" pt="xs">
                    <VideoSettings />
                </Tabs.Panel>
                <Tabs.Panel value="camera" pt="xs">
                    <CameraSettings />
                </Tabs.Panel>
                <Tabs.Panel value="alink" pt="xs">
                    <SystemSettings onAlinkChange={setAlinkEnabled} />
                </Tabs.Panel>
                <Tabs.Panel value="txprofiles" pt="xs">
                    <TxProfilesSettings />
                </Tabs.Panel>
            </Tabs>
        </Paper>
    );
}
