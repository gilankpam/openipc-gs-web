import { ActionIcon, Modal, Tooltip, Box } from '@mantine/core'; // Removed unused imports
import { useDisclosure, useFullscreen } from '@mantine/hooks';
import { IconWifi, IconWifiOff, IconSettings, IconMaximize, IconMinimize } from '@tabler/icons-react'; // Removed unused IconX
import { VideoPlayer } from './components/VideoPlayer';
import { ConfigPanel } from './components/ConfigPanel';
import { WFBStats } from './components/WFBStats';
import { useConnectionStatus } from './hooks/useConnectionStatus';

export default function App() {
  const [opened, { open, close }] = useDisclosure(false);
  const { toggle, fullscreen } = useFullscreen();
  const isConnected = useConnectionStatus();

  return (
    <div style={{ position: 'relative', width: '100vw', height: '100vh', overflow: 'hidden', backgroundColor: '#000' }}>
      {/* Video Background */}
      <VideoPlayer />

      {/* Connection Status - Top Left */}
      <Box style={{ position: 'absolute', top: 20, left: 20, zIndex: 100 }}>
        <Tooltip label={isConnected ? "Air Unit Connected" : "Air Unit Disconnected"} position="right">
          <div>
            {isConnected ? (
              <IconWifi size={32} color="lime" style={{ filter: 'drop-shadow(0px 0px 4px rgba(0,0,0,0.8))' }} />
            ) : (
              <IconWifiOff size={32} color="red" style={{ filter: 'drop-shadow(0px 0px 4px rgba(0,0,0,0.8))' }} />
            )}
          </div>
        </Tooltip>
      </Box>

      {/* WFB Stats - Bottom Left (handled by component absolute positioning) */}
      <WFBStats />

      {/* Settings Button - Top Right */}
      <Box style={{ position: 'absolute', top: 20, right: 20, zIndex: 100 }}>
        <ActionIcon
          onClick={open}
          variant="transparent"
          size="xl"
          aria-label="Settings"
          style={{ filter: 'drop-shadow(0px 0px 4px rgba(0,0,0,0.8))' }}
        >
          <IconSettings size={32} color="white" />
        </ActionIcon>
      </Box>

      {/* Fullscreen Button - Bottom Right */}
      <Box style={{ position: 'absolute', bottom: 20, right: 20, zIndex: 100 }}>
        <ActionIcon
          onClick={toggle}
          variant="transparent"
          size="xl"
          aria-label="Toggle Fullscreen"
          style={{ filter: 'drop-shadow(0px 0px 4px rgba(0,0,0,0.8))' }}
        >
          {fullscreen ? <IconMinimize size={32} color="white" /> : <IconMaximize size={32} color="white" />}
        </ActionIcon>
      </Box>

      {/* Settings Modal */}
      <Modal
        opened={opened}
        onClose={close}
        title="Settings"
        centered
        size="lg"
        zIndex={2000}
        overlayProps={{
          backgroundOpacity: 0.55,
          blur: 3,
        }}
      >
        <ConfigPanel isConnected={isConnected} />
      </Modal>
    </div>
  );
}
