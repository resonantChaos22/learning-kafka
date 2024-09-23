import WebSocketChart from "./components/Graph";

export default function Home() {
  return (
    <div className="mx-auto my-auto flex justify-center h-[100vh] items-center">
      <div className="text-8xl">
        <WebSocketChart />
      </div>
    </div>
  );
}
