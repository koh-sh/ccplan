import { describe, test, expect, afterEach } from "bun:test";
import {
  launchCommdPR,
  addComment,
  MOCK_PR_URL,
  TEST_TIMEOUT,
} from "../helpers/session";
import {
  createMockGitHubServer,
  type MockGitHubServer,
  type MockPRConfig,
} from "../helpers/mock-github";
import type { Session } from "tuistory";

// Markdown content served by the mock GitHub API.
const BASIC_MD = `# Plan: Authentication System

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
const BASIC_PATCH = `@@ -1,7 +1,7 @@
 # Plan: Authentication System

-This plan implements a basic authentication system.
+This plan implements a comprehensive authentication system.

 ## Step 1: Auth Middleware

`;

// Patch targeting Step 1 section (non-overview) with proper context empty lines.
// Uses " " (space-prefixed empty lines) so ParsePatch doesn't skip them.
const STEP1_PATCH = [
  "@@ -5,7 +5,7 @@",
  " ## Step 1: Auth Middleware",
  " ",
  "-Implement authentication middleware in `pkg/auth/middleware.go`.",
  "+Implement authentication middleware in `internal/auth/middleware.go`.",
  " ",
  " ### 1.1 JWT Verification",
  " ",
].join("\n");

const SECOND_MD = `# API Guide

This is the API documentation.

## Endpoints

Description of endpoints.
`;

function defaultConfig(): MockPRConfig {
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

// ──────────────────────────────────────────────────────────
// A. File Picker
// ──────────────────────────────────────────────────────────
describe("PR File Picker", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  test("shows changed MD files with Select title", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    const text = await session.waitForText("Select Markdown files", { timeout: 15000 });
    expect(text).toContain("docs/README.md");
    expect(text).toContain("docs/guide.md");
  }, TEST_TIMEOUT);

  test("all files selected by default", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    const text = await session.waitForText("Select Markdown files", { timeout: 15000 });
    // Count checkmarks - should match number of files
    const checks = text.match(/\[✓\]/g);
    expect(checks).not.toBeNull();
    expect(checks!.length).toBe(2);
  }, TEST_TIMEOUT);

  test("space toggles file selection", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    // Deselect first file
    await session.press(" ");
    const text = await session.text();
    expect(text).toContain("[ ]");
  }, TEST_TIMEOUT);

  test("a toggles all files selection", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    // All selected by default; press a to deselect all
    await session.press("a");
    let text = await session.text();
    expect(text).not.toContain("[✓]");
    // Press a again to select all
    await session.press("a");
    text = await session.text();
    const checks = text.match(/\[✓\]/g);
    expect(checks).not.toBeNull();
    expect(checks!.length).toBe(2);
  }, TEST_TIMEOUT);

  test("all files deselected then enter exits without review", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    // Deselect all
    await session.press("a");
    await session.press("enter");
    await session.waitIdle({ timeout: 5000 });
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("q cancels without review", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    await session.press("q");
    await session.waitIdle({ timeout: 5000 });
    // Process should have exited; no review submitted
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("enter confirms and transitions to review TUI", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    await session.press("enter");
    // Review TUI should appear with status bar
    const text = await session.waitForText("quit", { timeout: 15000 });
    expect(text).toContain("submit");
  }, TEST_TIMEOUT);
});

// ──────────────────────────────────────────────────────────
// B. --file flag
// ──────────────────────────────────────────────────────────
describe("PR --file flag", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  test("skips picker and shows review TUI directly", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    const text = await session.waitForText("quit", { timeout: 15000 });
    expect(text).toContain("submit");
    // Should NOT show file picker
    expect(text).not.toContain("Select Markdown files");
  }, TEST_TIMEOUT);

  test("nonexistent file shows error", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "missing.md",
    });
    await session.waitIdle({ timeout: 10000 });
    const text = await session.text({ immediate: true });
    expect(text).toContain("not found");
  }, TEST_TIMEOUT);
});

// ──────────────────────────────────────────────────────────
// C. PR Mode Review TUI
// ──────────────────────────────────────────────────────────
describe("PR Mode Review TUI", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  test("PR mode status bar shows submit and quit", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    const text = await session.waitForText("quit", { timeout: 15000 });
    expect(text).toContain("submit");
    expect(text).toContain("quit");
    expect(text).toContain("switch");
  }, TEST_TIMEOUT);

  test("diff view is default when patch exists", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    const text = await session.waitForText("quit", { timeout: 15000 });
    // In raw/diff mode, status bar shows "r render" to toggle TO render mode
    expect(text).toContain("render");
    // Diff content should show + and - lines
    expect(text).toContain("comprehensive");
  }, TEST_TIMEOUT);

  test("r toggles between render and raw view", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("render", { timeout: 15000 });
    // Switch to render mode
    await session.press("r");
    const rendered = await session.waitForText("raw", { timeout: 5000 });
    // Status bar now shows "r raw" to toggle back
    expect(rendered).toContain("raw");
    // Switch back to diff mode
    await session.press("r");
    const diffView = await session.waitForText("render", { timeout: 5000 });
    expect(diffView).toContain("render");
  }, TEST_TIMEOUT);

  test("section comment in render mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch to render mode (rawView off), focus stays on left pane
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    // Navigate to a section and add comment
    await session.press("j"); // Step 1
    await session.press("c");
    await session.waitForText("save");
    await session.type("section level feedback");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    const text = await session.text();
    expect(text).toContain("[*]");
    expect(text).toContain("section level feedback");
  }, TEST_TIMEOUT);

  test("line comment in diff view", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch focus to right pane (diff view)
    await session.press("tab");
    // Navigate down a few lines in diff
    await session.press("j");
    await session.press("j");
    // Add line comment
    await session.press("c");
    const commentText = await session.waitForText("save");
    // Should show line reference
    expect(commentText).toMatch(/\(L\d+\)/);
    await session.type("line level feedback");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    const text = await session.text();
    expect(text).toContain("line level feedback");
  }, TEST_TIMEOUT);

  test("range comment in diff view with visual select", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch focus to right pane
    await session.press("tab");
    // Start visual select
    await session.press("V");
    await session.waitForText("VISUAL");
    // Extend selection
    await session.press("j");
    // Comment on range
    await session.press("c");
    const commentText = await session.waitForText("save");
    // Should show range like (L1-L2)
    expect(commentText).toMatch(/\(L\d+-L\d+\)/);
    await session.type("range feedback");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    const text = await session.text();
    expect(text).toContain("range feedback");
  }, TEST_TIMEOUT);

  test("s shows PR mode confirm: Finish reviewing this file?", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("s");
    const text = await session.waitForText("Finish reviewing");
    expect(text).toContain("Finish reviewing this file?");
  }, TEST_TIMEOUT);

  test("c on left pane in raw view does nothing", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Default: rawView=true, focus=FocusLeft
    // Move to a section and try to comment
    await session.press("j");
    await session.press("c");
    // Should still be in normal mode (no comment editor opened)
    const text = await session.text();
    expect(text).not.toContain("save");
    expect(text).toContain("submit");
  }, TEST_TIMEOUT);

  test("edit existing comment in PR mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Add comment in render mode
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    await session.press("j");
    await session.press("c");
    await session.waitForText("save");
    await session.type("original comment");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Open comment list and edit
    await session.press("C");
    await session.waitForText("edit");
    await session.press("e");
    await session.waitForText("save");
    // Clear text: ctrl+e (end of line) then ctrl+u (kill to beginning)
    await session.press(["ctrl", "e"]);
    await session.press(["ctrl", "u"]);
    await session.type("edited comment");
    await session.press(["ctrl", "s"]);
    // After editing, returns to comment list mode; press esc to go back
    await session.waitForText("edit");
    await session.press("escape");
    await session.waitForText("quit");
    const text = await session.text();
    expect(text).toContain("edited comment");
  }, TEST_TIMEOUT);

  test("delete comment in PR mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Add comment
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    await session.press("j");
    await session.press("c");
    await session.waitForText("save");
    await session.type("to be deleted");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    let text = await session.text();
    expect(text).toContain("[*]");
    // Open comment list and delete
    await session.press("C");
    await session.waitForText("delete");
    await session.press("d");
    // Comment deleted → back to normal mode
    await session.waitForText("quit");
    text = await session.text();
    expect(text).not.toContain("[*]");
    expect(text).not.toContain("to be deleted");
  }, TEST_TIMEOUT);

  test("f toggles fullView in raw/diff mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Default raw view shows "f full" to switch to full view
    // Press f to toggle
    await session.press("f");
    // In full view, status bar shows "f section" to switch back
    const text = await session.text();
    expect(text).toContain("section");
  }, TEST_TIMEOUT);

  test("viewed mark toggle works in PR mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch to render mode so status bar shows viewed count
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    // Navigate to a section
    await session.press("j"); // Step 1
    // Toggle viewed mark
    await session.press("v");
    const text = await session.waitForText("[✓]", { timeout: 5000 });
    expect(text).toContain("1/");
    // Toggle again to unmark
    await session.press("v");
    const unmarked = await session.waitForText("0/", { timeout: 5000 });
    expect(unmarked).not.toContain("[✓]");
  }, TEST_TIMEOUT);

  test("search filters sections in PR mode", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch to render mode (search is on left pane which needs non-raw for /)
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    // Open search and type query (same pattern as existing search.test.ts)
    await session.press("/");
    await session.type("Routing");
    const text = await session.text({ trimEnd: true });
    expect(text).toContain("Routing");
    expect(text).not.toContain("Auth Middleware");
    // Esc restores all sections
    await session.press("escape");
    const cleared = await session.waitForText("Auth Middleware", { timeout: 5000 });
    expect(cleared).toContain("Routing");
  }, TEST_TIMEOUT);

  test("no changes in section shows message in diff view", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // BASIC_PATCH only covers lines 1-7 (overview + Step 1 area)
    // Navigate to Step 2 which is beyond the diff range
    await session.press("j"); // Step 1
    await session.press("j"); // 1.1
    await session.press("j"); // 1.2
    await session.press("j"); // Step 2
    const text = await session.text();
    expect(text).toContain("No changes in this section");
  }, TEST_TIMEOUT);

  test("multiple comments on same section", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch to render mode for section comments
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    // Add first comment on Step 1
    await session.press("j");
    await session.press("c");
    await session.waitForText("save");
    await session.type("first comment");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Add second comment on same section
    await session.press("c");
    await session.waitForText("save");
    await session.type("second comment");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Open comment list to verify both exist
    await session.press("C");
    const text = await session.waitForText("delete");
    expect(text).toContain("#1");
    expect(text).toContain("#2");
  }, TEST_TIMEOUT);

  test("raw source view toggle on file without patch", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/guide.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Added file without patch starts in render mode
    // Status bar should show "r raw" (toggle TO raw)
    let text = await session.text();
    expect(text).toContain("raw");
    // Press r to switch to raw source view
    await session.press("r");
    text = await session.waitForText("render", { timeout: 5000 });
    // Now in raw view, should show source lines
    expect(text).toContain("render");
    expect(text).toContain("# API Guide");
  }, TEST_TIMEOUT);

  test("q shows PR mode confirm: Skip this file?", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("q");
    const text = await session.waitForText("Skip this file?");
    expect(text).toContain("Skip this file?");
  }, TEST_TIMEOUT);
});

// ──────────────────────────────────────────────────────────
// D. Review Dialog
// ──────────────────────────────────────────────────────────
describe("Review Dialog", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  /** Helper: navigate through review TUI to ReviewDialog with a comment */
  async function goToReviewDialogWithComment(s: Session): Promise<void> {
    // Switch to render mode for section comment
    await s.press("r");
    await s.waitForText("raw", { timeout: 5000 });
    // Add comment on first real section
    await s.press("j");
    await s.press("c");
    await s.waitForText("save");
    await s.type("test comment");
    await s.press(["ctrl", "s"]);
    await s.waitForText("quit");
    // Submit
    await s.press("s");
    await s.waitForText("Finish reviewing");
    await s.press("y");
  }

  /** Helper: navigate through review TUI to ReviewDialog without comments */
  async function goToReviewDialogNoComment(s: Session): Promise<void> {
    await s.press("s");
    await s.waitForText("Finish reviewing");
    await s.press("y");
  }

  test("with comments shows Comment, Approve, Exit options", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    const text = await session.waitForText("cancel", { timeout: 15000 });
    expect(text).toContain("Comment");
    expect(text).toContain("Approve");
    expect(text).toContain("Exit");
  }, TEST_TIMEOUT);

  test("without comments shows Approve and Exit only", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogNoComment(session);
    const text = await session.waitForText("cancel", { timeout: 15000 });
    expect(text).toContain("Approve");
    expect(text).toContain("Exit");
    // "Comment" option should not be present when there are no comments
    // The dialog should show "No comments" in summary
    expect(text).toContain("No comments");
  }, TEST_TIMEOUT);

  test("Comment selection submits with event COMMENT", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // "Comment" is the first option (cursor is already there)
    await session.press("enter");
    // Body input mode
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    expect(mock.submittedReviews).toHaveLength(1);
    expect(mock.submittedReviews[0].event).toBe("COMMENT");
  }, TEST_TIMEOUT);

  test("Approve selection submits with event APPROVE", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // Move to Approve (second option)
    await session.press("j");
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    expect(mock.submittedReviews).toHaveLength(1);
    expect(mock.submittedReviews[0].event).toBe("APPROVE");
  }, TEST_TIMEOUT);

  test("body text is included in submitted review", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // Select Comment
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    // Type body text
    await session.type("Overall looks good");
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    expect(mock.submittedReviews).toHaveLength(1);
    expect(mock.submittedReviews[0].body).toContain("Overall looks good");
  }, TEST_TIMEOUT);

  test("Exit selection cancels without submitting", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // Move to Exit (third option)
    await session.press("j");
    await session.press("j");
    await session.press("enter");
    await session.waitIdle({ timeout: 5000 });
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("q in select mode cancels without submitting", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // Press q to cancel (different code path from selecting Exit)
    await session.press("q");
    await session.waitIdle({ timeout: 5000 });
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("esc in body input returns to action selection", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    await goToReviewDialogWithComment(session);
    await session.waitForText("cancel", { timeout: 15000 });
    // Select Comment → body mode
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    // Press esc to go back
    await session.press("escape");
    // Should be back in select mode with navigation help
    const text = await session.waitForText("navigate", { timeout: 5000 });
    expect(text).toContain("Comment");
    expect(text).toContain("Approve");
  }, TEST_TIMEOUT);
});

// ──────────────────────────────────────────────────────────
// E. Submit Verification
// ──────────────────────────────────────────────────────────
describe("Submit Verification", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  test("inline comments are included in submitted review", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Add section comment in render mode
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    await session.press("j"); // Step 1
    await session.press("c");
    await session.waitForText("save");
    await session.type("needs refactoring");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Submit
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    // ReviewDialog: select Comment → submit
    await session.waitForText("cancel", { timeout: 15000 });
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    // Verify submitted review
    expect(mock.submittedReviews).toHaveLength(1);
    const review = mock.submittedReviews[0];
    expect(review.event).toBe("COMMENT");
    expect(review.comments.length).toBeGreaterThanOrEqual(1);
    expect(review.comments[0].path).toBe("docs/README.md");
    expect(review.comments[0].body).toContain("needs refactoring");
  }, TEST_TIMEOUT);

  test("comment on removed line has side LEFT", async () => {
    // Use STEP1_PATCH which modifies lines within Step 1 section (non-overview)
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        {
          filename: "docs/README.md",
          status: "modified",
          patch: STEP1_PATCH,
          content: BASIC_MD,
        },
      ],
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Navigate to Step 1 (which has the diff changes)
    await session.press("j"); // Step 1
    // Switch focus to right pane (diff view)
    await session.press("tab");
    // Display lines for Step 1 range:
    //   [0] context "## Step 1: Auth Middleware"
    //   [1] context " " (empty)
    //   [2] removed "Implement ... pkg/auth..."    ← target
    //   [3] added "Implement ... internal/auth..."
    //   [4] context " " (empty)
    //   [5] context "### 1.1 JWT Verification"
    // Navigate to the removed line (index 2)
    await session.press("j");
    await session.press("j");
    // Add line comment on the removed line
    await session.press("c");
    await session.waitForText("save");
    await session.type("this line was removed");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Submit: s → y → ReviewDialog → Comment → ctrl+s
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    await session.waitForText("cancel", { timeout: 15000 });
    await session.press("enter"); // Comment
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    // Verify the comment has side=LEFT
    expect(mock.submittedReviews).toHaveLength(1);
    const review = mock.submittedReviews[0];
    expect(review.comments.length).toBeGreaterThanOrEqual(1);
    const comment = review.comments[0];
    expect(comment.side).toBe("LEFT");
    expect(comment.body).toContain("this line was removed");
  }, TEST_TIMEOUT);

  test("approve without comments submits empty review", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Submit without comments
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    // ReviewDialog: no comments → Approve is first option
    await session.waitForText("cancel", { timeout: 15000 });
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    expect(mock.submittedReviews).toHaveLength(1);
    expect(mock.submittedReviews[0].event).toBe("APPROVE");
  }, TEST_TIMEOUT);
});

// ──────────────────────────────────────────────────────────
// F. Edge Cases
// ──────────────────────────────────────────────────────────
describe("PR Edge Cases", () => {
  let session: Session;
  let mock: MockGitHubServer;

  afterEach(() => {
    session?.close();
    mock?.close();
  });

  test("no MD files shows message", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "main.go", status: "modified" },
      ],
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
    });
    await session.waitIdle({ timeout: 10000 });
    const text = await session.text({ immediate: true });
    expect(text).toContain("No Markdown files");
  }, TEST_TIMEOUT);

  test("all files skipped exits without ReviewDialog", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", patch: BASIC_PATCH, content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    // File picker
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    await session.press("enter");
    // Skip first file
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("q");
    await session.waitForText("Skip this file?");
    await session.press("y");
    // Skip second file
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("q");
    await session.waitForText("Skip this file?");
    await session.press("y");
    // Process should exit without ReviewDialog
    await session.waitIdle({ timeout: 5000 });
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("multi-file mixed flow: comment on file 1, skip file 2", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", patch: BASIC_PATCH, content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    // File picker: confirm all
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    await session.press("enter");
    // File 1: add comment and submit
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("r"); // render mode
    await session.waitForText("raw", { timeout: 5000 });
    await session.press("j");
    await session.press("c");
    await session.waitForText("save");
    await session.type("file 1 comment");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    // File 2: skip
    await session.waitForText("quit", { timeout: 15000 });
    await session.press("q");
    await session.waitForText("Skip this file?");
    await session.press("y");
    // ReviewDialog: only file 1 should appear in summary
    const dialogText = await session.waitForText("cancel", { timeout: 15000 });
    expect(dialogText).toContain("docs/README.md");
    expect(dialogText).not.toContain("docs/guide.md");
    // Submit as Comment
    await session.press("enter");
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    // Verify only file 1 comments submitted
    expect(mock.submittedReviews).toHaveLength(1);
    const review = mock.submittedReviews[0];
    expect(review.comments.length).toBeGreaterThanOrEqual(1);
    expect(review.comments[0].path).toBe("docs/README.md");
  }, TEST_TIMEOUT);

  test("added file without patch starts in render mode", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/guide.md",
    });
    const text = await session.waitForText("quit", { timeout: 15000 });
    // No diff → render mode is default → status bar should NOT show "render"
    // Instead it should show "raw" (toggle TO raw) or no raw toggle at all
    expect(text).toContain("submit");
    // Markdown content should be rendered (not diff lines)
    expect(text).toContain("API Guide");
  }, TEST_TIMEOUT);

  test("overview comment is filtered with warning on submit", async () => {
    mock = createMockGitHubServer(defaultConfig());
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Switch to render mode
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    // Overview is the first item (cursor starts there), add comment on it
    await session.press("c");
    await session.waitForText("save");
    await session.type("overview comment");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Submit
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    // ReviewDialog appears — overview comment only, so dialog shows "Comment" option
    await session.waitForText("cancel", { timeout: 15000 });
    await session.press("enter"); // Comment
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    // Overview comments are filtered out during submit.
    // With COMMENT event, empty comments and empty body → "No comments to submit"
    // So no review should be submitted to the mock.
    expect(mock.submittedReviews).toHaveLength(0);
  }, TEST_TIMEOUT);

  test("API submit error prints review content as fallback", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      reviewStatusCode: 422,
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitForText("quit", { timeout: 15000 });
    // Add comment
    await session.press("r");
    await session.waitForText("raw", { timeout: 5000 });
    await session.press("j");
    await session.press("c");
    await session.waitForText("save");
    await session.type("important feedback");
    await session.press(["ctrl", "s"]);
    await session.waitForText("quit");
    // Submit
    await session.press("s");
    await session.waitForText("Finish reviewing");
    await session.press("y");
    // ReviewDialog
    await session.waitForText("cancel", { timeout: 15000 });
    await session.press("enter"); // Comment
    await session.waitForText("back", { timeout: 5000 });
    await session.press(["ctrl", "s"]);
    await session.waitIdle({ timeout: 10000 });
    // API returned 422 → fallback prints review content to stderr
    const text = await session.text({ immediate: true });
    expect(text).toContain("Review content");
    expect(text).toContain("important feedback");
  }, TEST_TIMEOUT);

  test("GetHeadSHA API error shows error message", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      pullStatusCode: 404,
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
      file: "docs/README.md",
    });
    await session.waitIdle({ timeout: 10000 });
    const text = await session.text({ immediate: true });
    expect(text).toContain("404");
  }, TEST_TIMEOUT);

  test("ListMDFiles API error shows error message", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      filesStatusCode: 500,
    });
    session = await launchCommdPR({
      prURL: MOCK_PR_URL,
      mockServerURL: mock.url,
    });
    await session.waitIdle({ timeout: 10000 });
    const text = await session.text({ immediate: true });
    expect(text).toContain("500");
  }, TEST_TIMEOUT);

  test("skip first file proceeds to second file", async () => {
    mock = createMockGitHubServer({
      ...defaultConfig(),
      files: [
        { filename: "docs/README.md", status: "modified", patch: BASIC_PATCH, content: BASIC_MD },
        { filename: "docs/guide.md", status: "added", content: SECOND_MD },
      ],
    });
    session = await launchCommdPR({ prURL: MOCK_PR_URL, mockServerURL: mock.url });
    // File picker
    await session.waitForText("Select Markdown files", { timeout: 15000 });
    await session.press("enter"); // confirm all files
    // First file review TUI
    await session.waitForText("quit", { timeout: 15000 });
    // Skip first file
    await session.press("q");
    await session.waitForText("Skip this file?");
    await session.press("y");
    // Second file review TUI should appear
    const text = await session.waitForText("quit", { timeout: 15000 });
    expect(text).toContain("submit");
  }, TEST_TIMEOUT);
});
