import React, { useEffect, useState } from "react";
import { fetchHistory } from "./api";
import CommitCard from "./components/CommitCard";
import "./App.css";

function App() {
  const [commits, setCommits] = useState([]);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    const loadData = async () => {
      try {
        const data = await fetchHistory(10);
        setCommits(data);
      } catch (err) {
        console.error("Error fetching history:", err);
      } finally {
        setLoading(false);
      }
    };
    loadData();
  }, []);

  return (
    <div className="container">
      <h1>AutoGuard Dashboard</h1>
      {loading ? (
        <p>Loading commit history...</p>
      ) : (
        commits.map((commit, idx) => <CommitCard key={idx} commit={commit} />)
      )}
    </div>
  );
}

export default App;
