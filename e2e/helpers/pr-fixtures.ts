import type { MockPRConfig } from "./mock-github";

// Markdown content served by the mock GitHub API.
export const BASIC_MD = `# Plan: Authentication System

This plan implements a basic authentication system.

## Step 1: Auth Middleware

Implement authentication middleware in \`pkg/auth/middleware.go\`.

### 1.1 JWT Verification

Implement JWT token verification using RS256.

### 1.2 Middleware Registration

Register the middleware in the HTTP router.

## Step 2: Routing Updates

Update the routing configuration to use the new middleware.

### 2.1 Endpoint Addition

Add new protected endpoints.

### 2.2 Validation

Add request validation for auth endpoints.

## Step 3: Tests

Write comprehensive tests for the authentication system.
`;

// Unified diff patch for the markdown file.
export const BASIC_PATCH = `@@ -1,7 +1,7 @@
 # Plan: Authentication System

-This plan implements a basic authentication system.
+This plan implements a comprehensive authentication system.

 ## Step 1: Auth Middleware

`;

// Patch targeting Step 1 section (non-overview) with proper context empty lines.
// Uses " " (space-prefixed empty lines) so ParsePatch doesn't skip them.
export const STEP1_PATCH = [
  "@@ -5,7 +5,7 @@",
  " ## Step 1: Auth Middleware",
  " ",
  "-Implement authentication middleware in `pkg/auth/middleware.go`.",
  "+Implement authentication middleware in `internal/auth/middleware.go`.",
  " ",
  " ### 1.1 JWT Verification",
  " ",
].join("\n");

export const SECOND_MD = `# API Guide

This is the API documentation.

## Endpoints

Description of endpoints.
`;

export function defaultConfig(): MockPRConfig {
  return {
    owner: "test-owner",
    repo: "test-repo",
    number: 1,
    headSHA: "abc123",
    files: [
      {
        filename: "docs/README.md",
        status: "modified",
        patch: BASIC_PATCH,
        content: BASIC_MD,
      },
    ],
  };
}
