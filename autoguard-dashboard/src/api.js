import axios from "axios";

const API_BASE = "http://localhost:8080";
fetch("http://localhost:8080/history")
  .then(r => r.json())
  .then(console.log)

export const fetchHistory = async (limit = 10) => {
  const res = await axios.get(`${API_BASE}/history?limit=${limit}`);
  return res.data.commits;
};
