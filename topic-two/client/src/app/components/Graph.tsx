"use client";

import moment from "moment";
import React, { useEffect, useState, useRef } from "react";
import {
  Chart as ChartJS,
  LineElement,
  CategoryScale,
  LinearScale,
  PointElement,
  TimeScale,
  ChartData,
  TooltipItem,
  Tooltip,
  Legend,
  Filler,
  Title,
} from "chart.js";
import { Line } from "react-chartjs-2";
import "chartjs-adapter-moment";

ChartJS.register(
  LineElement,
  CategoryScale,
  LinearScale,
  PointElement,
  TimeScale,
  Tooltip,
  Filler,
  Title,
  Legend
);

type DataPoint = {
  time: number;
  value: number;
  id: number;
  ts: string;
};

type Coord = {
  x: moment.Moment;
  y: number;
};

const WebSocketChart: React.FC = () => {
  let timeRange = 1;
  const [data, setData] = useState<DataPoint[]>([]);
  const [chartData, setChartData] = useState<
    ChartData<"line", Coord[], moment.Moment>
  >({
    labels: [],
    datasets: [
      {
        label: "ItemA",
        data: [],
        fill: true,
        backgroundColor: "rgba(0, 255, 0, 0.1)",
        borderColor: "green",
        tension: 0.1,
      },
      {
        label: "ItemB",
        data: [],
        fill: true,
        backgroundColor: "rgba(128, 0, 128, 0.1)",
        borderColor: "purple",
        tension: 0.1,
      },
      {
        label: "ItemC",
        data: [],
        fill: true,
        backgroundColor: "rgba(0, 255, 255, 0.1)",
        borderColor: "cyan",
        tension: 0.1,
      },
    ],
  });
  const ws = useRef<WebSocket | null>(null);

  const generateLabels = (durationInMinutes: number) => {
    const now = moment();
    const labels = [];
    for (let i = 0; i <= durationInMinutes * 60; i += 10) {
      labels.push(moment(now).add(i, "seconds"));
    }
    return labels;
  };

  useEffect(() => {
    setChartData((prevData) => ({
      ...prevData,
      labels: generateLabels(timeRange),
    }));

    ws.current = new WebSocket("ws://localhost:8001/stream?itemID=0");

    ws.current.onmessage = (event) => {
      const newData: DataPoint = JSON.parse(event.data);
      const ts = moment(newData.time).format("HH:mm:ss");
      newData.ts = ts;
      setData((prevData) => [...prevData, newData]);
    };

    return () => {
      ws.current?.close();
    };
  }, []);

  useEffect(() => {
    if (data.length) {
      const newData = data[data.length - 1];

      setChartData((prevData) => {
        const duration = moment().diff(moment(prevData.labels?.[0]), "minutes");
        let labels;
        if (duration >= timeRange) {
          setTimeout(() => {
            timeRange++;
          }, 500);
          labels = generateLabels(timeRange);
        } else {
          labels = prevData.labels;
        }

        const updatedDatasets = prevData.datasets.map((dataset, index) => {
          if (newData.id === index + 1) {
            return {
              ...dataset,
              data: [
                ...dataset.data,
                { x: moment(newData.time), y: newData.value },
              ],
            };
          }
          return dataset;
        });

        return {
          ...prevData,
          labels,
          datasets: updatedDatasets,
        };
      });
    }
  }, [data]);

  return (
    <div>
      <Line
        width={1800}
        height={1200}
        data={chartData}
        title="Items Price Chart"
        options={{
          responsive: true,
          scales: {
            x: {
              type: "time",
              title: {
                display: true,
                color: "white",
                text: "Time",
                font: {
                  size: 20,
                },
              },
              time: {
                unit: "second",
                displayFormats: {
                  second: "HH:mm:ss",
                },
              },
              ticks: {
                color: "white",
              },
              grid: {
                color: "rgba(255, 255, 255, 0.1)",
              },
            },
            y: {
              beginAtZero: false,
              title: {
                display: true,
                color: "white",
                text: "Item Price(₹)",
                font: {
                  size: 20,
                },
              },
              ticks: {
                color: "white",
              },
              grid: {
                color: "rgba(255, 255, 255, 0.1)",
              },
            },
          },
          plugins: {
            tooltip: {
              enabled: true,
              callbacks: {
                label: (context: TooltipItem<"line">) => {
                  const label = context.dataset.label || "";
                  const value = context.raw as { x: moment.Moment; y: number };
                  return `${label}: ₹${value.y}`;
                },
              },
            },
            legend: {
              labels: {
                color: "white",
              },
            },
            title: {
              display: true,
              color: "white",
              text: "Items Price Chart",
              font: {
                size: 28,
              },
            },
          },
          layout: {
            padding: 10,
          },
        }}
      />
    </div>
  );
};

export default WebSocketChart;
