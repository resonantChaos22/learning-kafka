"use client";

import moment from "moment";
import React, { useEffect, useState, useRef } from "react";
import { LineChart, XAxis, YAxis, CartesianGrid, Line } from "recharts";
type DataPoint = {
  time: number;
  value: number;
  id: number;
  ts: string;
};

const WebSocketChart: React.FC = () => {
  const [data, setData] = useState<DataPoint[]>([]);
  const ws = useRef<WebSocket | null>(null);

  // Function to generate Y-axis ticks based on first value
  const generateYAxisTicks = (minValue: number) => {
    const ticks = [];
    const startTick = Math.floor(minValue / 25) * 25; // Round to the nearest 50 below
    for (let i = startTick - 25; i <= minValue + 80; i += 25) {
      ticks.push(i);
    }
    return ticks;
  };

  // Get current time and generate X-axis labels
  const currentTime = moment();
  const generateXAxisTicks = () => {
    return data.map((_, index) =>
      currentTime.clone().add(index, "minutes").format("HH:mm:ss")
    );
  };

  useEffect(() => {
    // Create WebSocket connection
    ws.current = new WebSocket("ws://localhost:8001/stream?itemID=1");

    ws.current.onmessage = (event) => {
      const newData: DataPoint = JSON.parse(event.data);
      const timeDamn = newData.time;
      const ts = moment(newData.time).format("HH:mm:ss");
      newData.ts = ts;
      console.log({ obj: newData, timeDamn, ts });
      setData((prevData) => [...prevData, newData]);
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  useEffect(() => {
    console.log(data);
  }, [data]);

  return (
    <div>
      <div>
        {/* {data.map((dataPoint: DataPoint) => (
          <div key={dataPoint.time}>
            <span>Timestamp: {moment(dataPoint.time).format("HH:mm:ss")}</span>
            <span> | Value: {dataPoint.value}</span>
          </div>
        ))} */}
        <LineChart width={1200} height={600} data={data}>
          <XAxis dataKey="ts" ticks={generateXAxisTicks()} />
          <YAxis
            domain={data[0] ? [data[0].value - 80, "auto"] : [500 - 80, "auto"]}
            ticks={generateYAxisTicks(data[0] ? data[0].value : 500)}
            tickCount={10}
          />
          <CartesianGrid stroke="#eee" strokeDasharray="5 5" />
          <Line type="monotone" dataKey="value" stroke="#8884d8" />
        </LineChart>
      </div>
      {/* Your LineChart or other UI components can be added here */}
    </div>
  );
};

export default WebSocketChart;
