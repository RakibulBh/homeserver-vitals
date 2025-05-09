"use client";

import { useEffect, useState } from "react";
import { LineChart, Line, XAxis, YAxis, ResponsiveContainer } from "recharts";
import { SystemVitals } from "@/types";
import { formatBytes, formatUptime } from "@/lib/utils";

export default function Home() {
  const [vitals, setVitals] = useState<SystemVitals | null>(null);
  const [cpuHistory, setCpuHistory] = useState<
    Array<{ time: string; value: number }>
  >([]);
  const [error, setError] = useState<string | null>(null);
  const [currentTime, setCurrentTime] = useState<string>("");

  useEffect(() => {
    // Set initial time and update it every second
    const updateTime = () => {
      setCurrentTime(new Date().toLocaleString());
    };

    updateTime();
    const timer = setInterval(updateTime, 1000);

    const eventSource = new EventSource(
      process.env.NEXT_PUBLIC_API_URL + "/sse"
    );

    eventSource.onmessage = (event) => {
      try {
        const data = JSON.parse(event.data) as SystemVitals;
        setVitals(data);

        // Update CPU history for the chart
        setCpuHistory((prev) => {
          const now = new Date().toLocaleTimeString();
          const newHistory = [...prev, { time: now, value: data.cpuUsage }];
          // Keep last 20 data points
          if (newHistory.length > 20) {
            return newHistory.slice(newHistory.length - 20);
          }
          return newHistory;
        });
      } catch (err) {
        console.error("Failed to parse SSE data:", err);
        console.log("Received data:", event.data);
        setError("Failed to parse server data");
      }
    };

    eventSource.onerror = (e) => {
      console.error("SSE connection error:", e);
      setError("Connection to server failed");
      eventSource.close();
    };

    return () => {
      eventSource.close();
      clearInterval(timer);
    };
  }, []);

  if (error) {
    return (
      <div className="min-h-screen bg-gray-900 text-white flex items-center justify-center">
        <div className="p-8 bg-red-900/30 rounded-md border border-red-700">
          <h1 className="text-2xl font-bold mb-4">Error</h1>
          <p>{error}</p>
          <p className="mt-4">
            Please ensure the server is running at http://localhost:2000
          </p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-screen bg-[#1a1a2e] text-slate-100 p-4 md:p-6">
      <header className="mb-6">
        <h1 className="text-2xl md:text-3xl font-semibold text-white">
          System Vitals Dashboard
        </h1>
        {vitals?.hostInfo && (
          <div className="text-sm text-purple-300 mt-1">
            {vitals.hostInfo.hostname} · {vitals.hostInfo.platform}{" "}
            {vitals.hostInfo.platformVersion}
          </div>
        )}
      </header>

      {!vitals ? (
        <div className="flex items-center justify-center h-[80vh]">
          <div className="animate-pulse flex flex-col items-center">
            <div className="h-6 w-32 bg-purple-800/50 rounded mb-4"></div>
            <div className="h-4 w-64 bg-purple-800/30 rounded"></div>
          </div>
        </div>
      ) : (
        <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-4 md:gap-6">
          {/* CPU Usage with Chart */}
          <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <div className="flex justify-between items-center mb-4">
              <h2 className="text-lg font-medium text-white">CPU Usage</h2>
              <span className="text-2xl font-semibold text-white">
                {vitals.cpuUsage.toFixed(1)}%
              </span>
            </div>
            <div className="h-[150px]">
              <ResponsiveContainer width="100%" height="100%">
                <LineChart data={cpuHistory}>
                  <XAxis
                    dataKey="time"
                    tick={{ fill: "#a4a8be", fontSize: 10 }}
                    axisLine={{ stroke: "#393e5c" }}
                    tickLine={{ stroke: "#393e5c" }}
                  />
                  <YAxis
                    domain={[0, 100]}
                    tick={{ fill: "#a4a8be", fontSize: 10 }}
                    axisLine={{ stroke: "#393e5c" }}
                    tickLine={{ stroke: "#393e5c" }}
                  />
                  <Line
                    type="monotone"
                    dataKey="value"
                    stroke="#7e57c2"
                    strokeWidth={2}
                    dot={false}
                    activeDot={{ r: 4, fill: "#9575cd" }}
                  />
                </LineChart>
              </ResponsiveContainer>
            </div>
          </div>

          {/* Memory Usage */}
          <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <h2 className="text-lg font-medium text-white mb-4">
              Memory Usage
            </h2>
            <div className="flex justify-between mb-2">
              <span className="text-slate-300">Used</span>
              <span className="text-white font-medium">
                {formatBytes(vitals.memory.used)}
              </span>
            </div>
            <div className="relative w-full h-2 bg-[#272b45] rounded-sm overflow-hidden mb-4">
              <div
                className="absolute top-0 left-0 h-full bg-purple-600"
                style={{ width: `${vitals.memory.usedPercent}%` }}
              ></div>
            </div>
            <div className="flex justify-between mb-1">
              <span className="text-slate-300">Total</span>
              <span className="text-slate-300">
                {formatBytes(vitals.memory.total)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-300">Usage</span>
              <span className="text-white font-medium">
                {vitals.memory.usedPercent.toFixed(1)}%
              </span>
            </div>
          </div>

          {/* Disk Usage */}
          {/* <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <h2 className="text-lg font-medium text-white mb-4">Disk Usage</h2>
            <div className="flex justify-between mb-2">
              <span className="text-slate-300">Used</span>
              <span className="text-white font-medium">
                {formatBytes(vitals.disk.used)}
              </span>
            </div>
            <div className="relative w-full h-2 bg-[#272b45] rounded-sm overflow-hidden mb-4">
              <div
                className="absolute top-0 left-0 h-full bg-[#5e60ce]"
                style={{ width: `${vitals.disk.usedPercent}%` }}
              ></div>
            </div>
            <div className="flex justify-between mb-1">
              <span className="text-slate-300">Total</span>
              <span className="text-slate-300">
                {formatBytes(vitals.disk.total)}
              </span>
            </div>
            <div className="flex justify-between">
              <span className="text-slate-300">Usage</span>
              <span className="text-white font-medium">
                {vitals.disk.usedPercent.toFixed(1)}%
              </span>
            </div>
          </div> */}

          {/* Network */}
          <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <h2 className="text-lg font-medium text-white mb-4">Network I/O</h2>
            <div className="grid grid-cols-2 gap-4">
              <div>
                <div className="flex items-center mb-1">
                  <div className="w-3 h-3 rounded-sm bg-purple-400 mr-2"></div>
                  <span className="text-slate-300">Download</span>
                </div>
                <div className="text-xl font-medium text-white">
                  {formatBytes(vitals.network.bytesRecv)}
                </div>
              </div>
              <div>
                <div className="flex items-center mb-1">
                  <div className="w-3 h-3 rounded-sm bg-indigo-400 mr-2"></div>
                  <span className="text-slate-300">Upload</span>
                </div>
                <div className="text-xl font-medium text-white">
                  {formatBytes(vitals.network.bytesSent)}
                </div>
              </div>
            </div>
          </div>

          {/* System Load */}
          <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <h2 className="text-lg font-medium text-white mb-4">System Load</h2>
            <div className="grid grid-cols-3 gap-2">
              <div>
                <div className="text-slate-300 text-sm mb-1">1 min</div>
                <div className="text-lg font-medium text-white">
                  {vitals.loadAvg.load1.toFixed(2)}
                </div>
              </div>
              <div>
                <div className="text-slate-300 text-sm mb-1">5 min</div>
                <div className="text-lg font-medium text-white">
                  {vitals.loadAvg.load5.toFixed(2)}
                </div>
              </div>
              <div>
                <div className="text-slate-300 text-sm mb-1">15 min</div>
                <div className="text-lg font-medium text-white">
                  {vitals.loadAvg.load15.toFixed(2)}
                </div>
              </div>
            </div>
          </div>

          {/* System Info */}
          <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
            <h2 className="text-lg font-medium text-white mb-4">
              System Information
            </h2>
            <div className="grid grid-cols-2 gap-x-4 gap-y-3">
              <div>
                <div className="text-slate-300 text-sm">Uptime</div>
                <div className="text-white font-medium">
                  {formatUptime(vitals.uptime)}
                </div>
              </div>
              <div>
                <div className="text-slate-300 text-sm">Processes</div>
                <div className="text-white font-medium">
                  {vitals.processes.toLocaleString()}
                </div>
              </div>
              <div>
                <div className="text-slate-300 text-sm">Go Routines</div>
                <div className="text-white font-medium">
                  {vitals.goRoutines.toLocaleString()}
                </div>
              </div>
              <div>
                <div className="text-slate-300 text-sm">Go Memory</div>
                <div className="text-white font-medium">
                  {formatBytes(vitals.goMemAlloc)}
                </div>
              </div>
            </div>
          </div>

          {/* Temperature */}
          {vitals.temperature && vitals.temperature.length > 0 && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">
                Temperature
              </h2>
              <div className="space-y-3">
                {vitals.temperature.slice(0, 4).map((temp, index) => (
                  <div
                    key={index}
                    className="flex justify-between items-center"
                  >
                    <span className="text-slate-300">{temp.sensorKey}</span>
                    <span
                      className={`font-medium ${
                        temp.temperature > 80
                          ? "text-red-400"
                          : temp.temperature > 60
                          ? "text-orange-400"
                          : "text-green-400"
                      }`}
                    >
                      {temp.temperature !== undefined &&
                      temp.temperature !== null
                        ? `${temp.temperature.toFixed(1)}°C`
                        : "N/A"}
                    </span>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* Hardware Information */}
          {vitals.hardware && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">Hardware</h2>
              <div className="space-y-2">
                <div>
                  <div className="text-slate-300 text-sm">CPU</div>
                  <div className="text-white font-medium">
                    {vitals.hardware.cpuModel}
                  </div>
                </div>
                <div className="flex justify-between">
                  <div>
                    <div className="text-slate-300 text-sm">Cores</div>
                    <div className="text-white">{vitals.hardware.cpuCores}</div>
                  </div>
                  <div>
                    <div className="text-slate-300 text-sm">Threads</div>
                    <div className="text-white">
                      {vitals.hardware.cpuThreads}
                    </div>
                  </div>
                </div>
                {vitals.hardware.systemModel !== "Unknown" && (
                  <div>
                    <div className="text-slate-300 text-sm">System</div>
                    <div className="text-white font-medium">
                      {vitals.hardware.systemVendor}{" "}
                      {vitals.hardware.systemModel}
                    </div>
                  </div>
                )}
              </div>
            </div>
          )}

          {/* Top Processes */}
          {vitals.topProcesses && vitals.topProcesses.length > 0 && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">
                Top Processes
              </h2>
              <div className="overflow-hidden">
                <table className="w-full text-sm">
                  <thead className="text-slate-300 border-b border-[#393e5c]">
                    <tr>
                      <th className="text-left pb-2">Process</th>
                      <th className="text-right pb-2">CPU</th>
                      <th className="text-right pb-2">Mem</th>
                    </tr>
                  </thead>
                  <tbody>
                    {vitals.topProcesses.map((process, i) => (
                      <tr
                        key={i}
                        className="border-b border-[#232b4a] last:border-0"
                      >
                        <td
                          className="py-2 text-white truncate"
                          title={process.command}
                        >
                          {process.name}
                        </td>
                        <td className="py-2 text-right text-purple-300">
                          {process.cpu.toFixed(1)}%
                        </td>
                        <td className="py-2 text-right text-indigo-300">
                          {process.memory.toFixed(1)}%
                        </td>
                      </tr>
                    ))}
                  </tbody>
                </table>
              </div>
            </div>
          )}

          {/* Network Interfaces */}
          {vitals.networkIfaces && vitals.networkIfaces.length > 0 && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">
                Network Interfaces
              </h2>
              <div className="space-y-3">
                {vitals.networkIfaces
                  .filter((iface) => iface.isUp)
                  .map((iface, i) => (
                    <div
                      key={i}
                      className="border-b border-[#232b4a] last:border-0 pb-2 last:pb-0"
                    >
                      <div className="flex justify-between mb-1">
                        <span className="text-white font-medium">
                          {iface.name}
                        </span>
                        <span
                          className={`text-xs px-1.5 py-0.5 rounded ${
                            iface.isUp
                              ? "bg-green-900/40 text-green-300"
                              : "bg-red-900/40 text-red-300"
                          }`}
                        >
                          {iface.isUp ? "UP" : "DOWN"}
                        </span>
                      </div>
                      <div className="text-sm text-purple-200 mb-1">
                        {iface.ipAddress || "No IP Address"}
                      </div>
                      <div className="text-xs text-slate-400">
                        MAC: {iface.macAddr || "Unknown"}
                      </div>
                    </div>
                  ))}
              </div>
            </div>
          )}

          {/* Disk Usage */}
          {vitals.disks && vitals.disks.length > 0 && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">Storage</h2>
              <div className="space-y-4">
                {vitals.disks.map((disk, index) => (
                  <div
                    key={index}
                    className="border-b border-[#232b4a] last:border-0 pb-3 last:pb-0"
                  >
                    <div className="flex justify-between mb-1">
                      <span className="text-slate-300 font-medium">
                        {disk.mountPoint}
                      </span>
                      <span className="text-xs text-slate-400">
                        {disk.fileSystem}
                      </span>
                    </div>
                    <div className="relative w-full h-2 bg-[#272b45] rounded-sm overflow-hidden mb-2">
                      <div
                        className="absolute top-0 left-0 h-full bg-[#5e60ce]"
                        style={{ width: `${disk.usedPercent}%` }}
                      ></div>
                    </div>
                    <div className="flex justify-between text-xs">
                      <span className="text-slate-300">
                        {formatBytes(disk.used)} / {formatBytes(disk.total)}
                      </span>
                      <span className="text-white font-medium">
                        {disk.usedPercent.toFixed(1)}%
                      </span>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* CPU Per Core */}
          {vitals.cpuPerCore && vitals.cpuPerCore.length > 0 && (
            <div className="bg-[#16213e] border border-[#393e5c] p-4 rounded-sm shadow-lg">
              <h2 className="text-lg font-medium text-white mb-4">CPU Cores</h2>
              <div className="space-y-3">
                {vitals.cpuPerCore.map((usage, i) => (
                  <div key={i}>
                    <div className="flex justify-between mb-1 text-sm">
                      <span className="text-slate-300">Core {i}</span>
                      <span className="text-white">{usage.toFixed(1)}%</span>
                    </div>
                    <div className="relative w-full h-1.5 bg-[#272b45] rounded-sm overflow-hidden">
                      <div
                        className="absolute top-0 left-0 h-full bg-gradient-to-r from-purple-600 to-indigo-600"
                        style={{ width: `${usage}%` }}
                      ></div>
                    </div>
                  </div>
                ))}
              </div>
            </div>
          )}

          {/* System Updates */}
          {vitals.systemUpdates > 0 && (
            <div className="bg-[#16213e] border border-amber-700/30 p-4 rounded-sm shadow-lg">
              <div className="flex items-center">
                <div className="w-8 h-8 flex items-center justify-center bg-amber-600/20 rounded-sm mr-3">
                  <svg
                    xmlns="http://www.w3.org/2000/svg"
                    className="h-5 w-5 text-amber-400"
                    viewBox="0 0 20 20"
                    fill="currentColor"
                  >
                    <path
                      fillRule="evenodd"
                      d="M10 18a8 8 0 100-16 8 8 0 000 16zm1-11a1 1 0 10-2 0v3.586L7.707 9.293a1 1 0 00-1.414 1.414l3 3a1 1 0 001.414 0l3-3a1 1 0 00-1.414-1.414L11 10.586V7z"
                      clipRule="evenodd"
                    />
                  </svg>
                </div>
                <div>
                  <h2 className="text-lg font-medium text-white">
                    System Updates Available
                  </h2>
                  <p className="text-amber-300">
                    {vitals.systemUpdates} package
                    {vitals.systemUpdates !== 1 ? "s" : ""} can be updated
                  </p>
                </div>
              </div>
            </div>
          )}
        </div>
      )}

      <footer className="mt-8 pt-4 border-t border-[#393e5c] text-slate-400 text-xs">
        <div className="flex justify-between items-center">
          <div>{currentTime}</div>
          <div>Server: http://localhost:2000</div>
        </div>
      </footer>
    </div>
  );
}
