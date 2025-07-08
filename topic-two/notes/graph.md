# FE Notes

## How does Graph FE work?

The provided code is a React component that renders a real-time line chart using **Chart.js**, a popular JavaScript library for data visualization. The chart dynamically updates as new data points are received, providing a visual representation of items' prices over time.

Here's a breakdown of how the code works, with a focus on the integration and configuration of Chart.js within the React component.

### 1. Importing Necessary Modules and Components

```jsx
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
```

- **Chart.js Modules**: The code imports various components from Chart.js that are essential for rendering a line chart, such as `LineElement`, `CategoryScale`, `LinearScale`, etc.
- **Moment.js**: Used for handling dates and times, crucial for plotting data over a time scale.
- **React Hooks**: `useEffect`, `useState`, and `useRef` are used for managing state and lifecycle methods within the React component.

### 2. Registering Chart.js Components

```jsx
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
```

- **Registration**: Chart.js requires you to register the components you intend to use. This modular approach allows for smaller bundle sizes by including only necessary components.

### 3. Defining Data Structures

```jsx
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
```

- **DataPoint**: Represents an individual data point received, containing a timestamp, value, identifier, and a formatted timestamp string.
- **Coord**: Represents a coordinate on the chart, with `x` as a moment object (time) and `y` as the value.

### 4. Setting Up State Variables

```jsx
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
```

- **Data State**: Holds the array of `DataPoint` objects received.
- **Chart Data State**: Configures the data and appearance for the Chart.js line chart.
  - **Labels**: Initially empty; will be populated with time values.
  - **Datasets**: Contains three datasets for "ItemA", "ItemB", and "ItemC", each with their own styling.

### 5. Generating Labels for the X-Axis

```jsx
const generateLabels = (durationInMinutes: number) => {
  const now = moment();
  const labels = [];
  for (let i = 0; i <= durationInMinutes * 60; i += 10) {
    labels.push(moment(now).add(i, "seconds"));
  }
  return labels;
};
```

- **Label Generation**: Creates an array of time labels at 10-second intervals over the specified duration.
- **Usage**: These labels are used for the x-axis of the chart to represent time.

### 6. Handling Component Mounting with useEffect

```jsx
useEffect(() => {
  setChartData((prevData) => ({
    ...prevData,
    labels: generateLabels(timeRange),
  }));

  // WebSocket connection setup (omitted as per instruction)

  return () => {
    // Cleanup on component unmount (omitted as per instruction)
  };
}, []);
```

- **Initial Label Setup**: On component mount, the `chartData` state is updated to include the generated labels.
- **Note**: The WebSocket setup is part of data retrieval and is not the focus here.

### 7. Updating the Chart When New Data Arrives

```jsx
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
```

- **Data Check**: Ensures there's new data to process.
- **Dynamic Label Adjustment**:
  - **Duration Calculation**: Determines if the current time range needs to be extended.
  - **Time Range Extension**: If the chart has reached the current time range limit, it extends the time range and regenerates labels.
- **Updating Datasets**:
  - **Dataset Mapping**: Iterates over existing datasets to find the matching dataset based on `id`.
  - **Data Appending**: Adds the new data point to the correct dataset.
- **State Update**: Updates the `chartData` state with new labels and datasets, triggering a re-render of the chart.

### 8. Rendering the Chart Component

```jsx
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
```

- **Line Component**: Renders the chart using the `Line` component from `react-chartjs-2`, which is a React wrapper for Chart.js.
- **Chart Dimensions**: Specifies the width and height of the chart.
- **Data Prop**: Passes the `chartData` state, which contains labels and datasets, to the chart.
- **Options Prop**: Customizes the appearance and behavior of the chart.
  - **Responsive Design**: Ensures the chart adjusts to different screen sizes.
  - **Scales Configuration**:
    - **X-Axis**:
      - **Type**: Set to `"time"` to handle time-series data.
      - **Title and Styling**: Customizes the axis title and tick appearance.
      - **Time Options**: Defines the unit and display format for time labels.
    - **Y-Axis**:
      - **Begin at Zero**: Set to `false` to adjust the y-axis based on data range.
      - **Title and Styling**: Similar customization as the x-axis.
  - **Plugins**:
    - **Tooltip**:
      - **Enabled**: Allows tooltips to appear on hover.
      - **Callbacks**: Customizes the tooltip label to display the item name and price.
    - **Legend**: Customizes the appearance of the legend labels.
    - **Title**: Sets the chart title and styling.
  - **Layout**: Adds padding around the chart.

### 9. Exporting the Component

```jsx
export default WebSocketChart;
```

- **Export**: Makes the `WebSocketChart` component available for import in other parts of the application.

---

### Chart.js Integration Highlights

- **Modular Imports**: Only necessary Chart.js components are imported and registered, optimizing performance.
- **Time Scale**: The x-axis is configured as a time scale, leveraging `chartjs-adapter-moment` for seamless date and time handling.
- **Dynamic Data Updates**: The chart data updates dynamically as new data points are added, showcasing Chart.js's ability to handle real-time data.
- **Styling and Customization**: Extensive customization options are used to style the chart, including colors, fonts, and grid lines.
- **Responsive Design**: The chart is set to be responsive, adapting to different screen sizes and resolutions.
- **Interactivity**: Tooltips and legends enhance user interaction, providing additional information upon hover and identifying different datasets.

---

**Note**: While the code includes WebSocket connections to receive real-time data, the explanation focuses on how the chart is set up and updated using Chart.js within a React component.
