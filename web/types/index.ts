// Type definitions based on Go structs
export type SystemVitals = {
  cpuUsage: number;
  cpuPerCore: number[];
  memory: {
    total: number;
    used: number;
    usedPercent: number;
  };
  swap?: {
    total: number;
    used: number;
    usedPercent: number;
  };
  disks: Array<{
    mountPoint: string;
    fileSystem: string;
    total: number;
    used: number;
    free: number;
    usedPercent: number;
  }>;
  network: {
    bytesSent: number;
    bytesRecv: number;
  };
  networkIfaces: Array<{
    name: string;
    ipAddress: string;
    macAddr: string;
    bytesSent: number;
    bytesRecv: number;
    isUp: boolean;
  }>;
  hostInfo: {
    hostname: string;
    platform: string;
    platformVersion: string;
  };
  uptime: number;
  loadAvg: {
    load1: number;
    load5: number;
    load15: number;
  };
  processes: number;
  temperature: Array<{
    sensorKey: string;
    temperature: number;
  }>;
  goRoutines: number;
  goMemAlloc: number;
  topProcesses: Array<{
    pid: number;
    name: string;
    cpu: number;
    memory: number;
    command: string;
  }>;
  hardware: {
    cpuModel: string;
    cpuCores: number;
    cpuThreads: number;
    totalMemory: number;
    systemVendor: string;
    systemModel: string;
  };
  lastUpdated: string;
  systemUpdates: number;
  diskIO: Record<
    string,
    {
      readCount: number;
      writeCount: number;
      readBytes: number;
      writeBytes: number;
      readTime: number;
      writeTime: number;
    }
  >;
};
