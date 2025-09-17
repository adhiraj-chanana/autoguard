import React from "react";
import "../App.css";

const CommitCard = ({ commit }) => {
  return (
    <div className="commit-card">
      <h2>
        Commit {commit.commit_id}{" "}
        <span style={{ color: commit.status === "pass" ? "green" : "red" }}>
          ({commit.status})
        </span>
      </h2>
      <div className="commit-meta">
        Repo: {commit.repo_url} <br />
        Timestamp: {commit.timestamp}
      </div>

      <div>
        {commit.issues.length > 0 ? (
          commit.issues.map((issue, idx) => (
            <div key={idx} className="issue">
              <p><b>Type:</b> {issue.type}</p>
              <p><b>Message:</b> {issue.message}</p>
              <p><b>File:</b> {issue.filename} (line {issue.line})</p>
              <p><b>Retries:</b> {issue.retries}</p>
            </div>
          ))
        ) : (
          <p className="success">✅ No issues</p>
        )}
      </div>
    </div>
  );
};

export default CommitCard;
