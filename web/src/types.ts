export interface RadioSettings {
    channel: number;
    bandwidth: number;
    tx_power: number;
    mcs_index: number;
    stbc: number;
    ldpc: number;
    fec_k: number;
    fec_n: number;
}

export interface VideoSettings {
    resolution: string;
    fps: number;
    bitrate: number;
    codec: 'h264' | 'h265';
    gop_size: number;
}

export interface CameraSettings {
    contrast: number;
    saturation: number;
    flip: boolean;
    mirror: boolean;
    rotate: number;
}

export interface TxProfile {
    range_start: number;
    range_end: number;
    gi: 'long' | 'short';
    mcs: number; // 0-7
    fec_k: number;
    fec_n: number;
    bitrate: number;
    gop: number;
    pwr: number;
    roi_qp: string;
    bandwidth: number; // 10, 20, 40
    qp_delta: number;
}

export interface AdaptiveLinkSettings {
    enabled: boolean;
    allow_set_power: boolean;
    power_level_0_to_4: number;
    allow_spike_fix_fps: boolean;
    osd_level: number;
}


