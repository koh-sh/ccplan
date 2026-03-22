import type { Server } from "bun";

export interface MockPRFile {
  filename: string;
  status: string;
  patch?: string;
  content?: string; // raw markdown content (will be base64-encoded in response)
}

export interface MockPRConfig {
  owner: string;
  repo: string;
  number: number;
  headSHA: string;
  files: MockPRFile[];
  pullStatusCode?: number; // if set, GET /pulls/{n} returns this status
  filesStatusCode?: number; // if set, GET /pulls/{n}/files returns this status
  reviewStatusCode?: number; // if set, POST /reviews returns this status instead of 200
}

export interface MockGitHubServer {
  server: Server;
  url: string;
  submittedReviews: any[];
  close: () => void;
}

/**
 * Create a mock GitHub API server using Bun.serve().
 * Port 0 lets the OS pick an available port (safe for CI parallel execution).
 */
export function createMockGitHubServer(
  config: MockPRConfig,
): MockGitHubServer {
  const submittedReviews: any[] = [];
  const { owner, repo, number } = config;
  const prefix = `/repos/${owner}/${repo}`;

  const server = Bun.serve({
    port: 0,
    async fetch(req) {
      const url = new URL(req.url);
      const path = url.pathname;

      // GET /repos/{owner}/{repo}/pulls/{number}
      if (
        req.method === "GET" &&
        path === `${prefix}/pulls/${number}`
      ) {
        if (config.pullStatusCode) {
          return Response.json(
            { message: "Not Found" },
            { status: config.pullStatusCode },
          );
        }
        return Response.json({
          head: { sha: config.headSHA, ref: "feature" },
        });
      }

      // GET /repos/{owner}/{repo}/pulls/{number}/files
      if (
        req.method === "GET" &&
        path === `${prefix}/pulls/${number}/files`
      ) {
        if (config.filesStatusCode) {
          return Response.json(
            { message: "Internal Server Error" },
            { status: config.filesStatusCode },
          );
        }
        return Response.json(
          config.files.map((f) => ({
            filename: f.filename,
            status: f.status,
            patch: f.patch ?? "",
          })),
        );
      }

      // GET /repos/{owner}/{repo}/contents/{filePath}
      if (req.method === "GET" && path.startsWith(`${prefix}/contents/`)) {
        const filePath = path.replace(`${prefix}/contents/`, "");
        const file = config.files.find((f) => f.filename === filePath);
        if (!file || !file.content) {
          return new Response("Not Found", { status: 404 });
        }
        const encoded = Buffer.from(file.content).toString("base64");
        return Response.json({
          type: "file",
          encoding: "base64",
          content: encoded,
        });
      }

      // POST /repos/{owner}/{repo}/pulls/{number}/reviews
      if (
        req.method === "POST" &&
        path === `${prefix}/pulls/${number}/reviews`
      ) {
        const body = await req.json();
        submittedReviews.push(body);
        if (config.reviewStatusCode) {
          return Response.json(
            { message: "Validation Failed" },
            { status: config.reviewStatusCode },
          );
        }
        return Response.json({ id: 1 });
      }

      return new Response("Not Found", { status: 404 });
    },
  });

  return {
    server,
    url: `http://localhost:${server.port}/`,
    submittedReviews,
    close: () => server.stop(),
  };
}
