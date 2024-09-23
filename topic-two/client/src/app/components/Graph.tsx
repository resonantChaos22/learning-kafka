"use client";

import moment from "moment";
import React, { useEffect, useState, useRef } from "react";
import { Chart as ChartJS, LineElement, CategoryScale, LinearScale, PointElement, TimeScale, ChartData, TimeSeriesScale } from 'chart.js';
import { Line } from "react-chartjs-2";
import 'chartjs-adapter-moment'

ChartJS.register(LineElement, CategoryScale, LinearScale, PointElement, TimeScale, TimeSeriesScale);
type DataPoint = {
  time: number;
  value: number;
  id: number;
  ts: string;
};

const WebSocketChart: React.FC = () => {
  const [data, setData] = useState<DataPoint[]>([]);
  const [chartData, setChartData] = useState<ChartData<'line', number[], string>>({
    labels: [],
    datasets: [{
      label: 'Value Over Time',
      data: [],
      fill: false,
      borderColor: 'white',
      tension: 0.1
    }]
  })
  const ws = useRef<WebSocket | null>(null);

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
    console.log({data})
    let newData: DataPoint;
    if (data.length) {
      newData = data[data.length - 1];
    } else {
      newData = {
        value: 500,
        ts: moment().format("HH:mm:ss"),
        time: moment().unix(),
        id: 1
      }
    }
    setChartData((prevData) => {
      const newLabels = [...(prevData.labels || []), moment(newData.time).format("HH:mm:ss")];
      const newValues = [...prevData.datasets[0].data, newData.value];

      return {
        ...prevData,
        labels: newLabels,
        datasets: [
          {
            ...prevData.datasets[0],
            data: newValues
          }
        ]
      }
    })
  }, [data]);

  useEffect(() => {
    console.log({chartData})
  }, [chartData])

  return (
    <div>
      <div>
        {/* {data.map((dataPoint: DataPoint) => (
          <div key={dataPoint.time}>
            <span>Timestamp: {moment(dataPoint.time).format("HH:mm:ss")}</span>
            <span> | Value: {dataPoint.value}</span>
          </div>
        ))} */}
        <Line width={1000} height={800} data={chartData} options={{
          scales: {
            x: {
              type: "timeseries",
              time: {
                unit: "second"
              }
            },
            y: {
              beginAtZero: false
            }
          }
        }} />
      </div>
      {/* Your LineChart or other UI components can be added here */}
    </div>
  );
};

export default WebSocketChart;
