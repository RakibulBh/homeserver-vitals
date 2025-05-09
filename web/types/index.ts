// Type definitions based on Go structs
export type SystemVitals = {
  cpuUsage: number;
  memory: {
    total: number;
    used: number;
    usedPercent: number;
  };
  disk: {
    total: number;
    used: number;
    usedPercent: number;
  };
  network: {
    bytesSent: number;
    bytesRecv: number;
  };
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
};
