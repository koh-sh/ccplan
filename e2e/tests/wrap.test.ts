import { describe, test, expect, afterEach } from "bun:test";
import { launchCommd, FIXTURE_CJK_WRAP, TEST_TIMEOUT } from "../helpers/session";
import type { Session } from "tuistory";

describe("CJK Wrap", () => {
  let session: Session;

  afterEach(() => {
    session?.close();
  });

  test("long Japanese paragraph wraps within the detail pane", async () => {
    session = await launchCommd({ file: FIXTURE_CJK_WRAP });
    await session.waitForText("これは日本語の長い段落です");
    const text = await session.text({ trimEnd: true });
    // If the paragraph were not wrapped, its tail would be cut off at the
    // right edge of the terminal and never appear on screen.
    expect(text).toContain("最後の文はここで終わります。");
    expect(text).toContain("意図的に長くしています。");
    // The paragraph must span multiple lines, not a single overflowing line.
    const lines = text.split("\n");
    const paragraphLines = lines.filter((l) =>
      /日本語の長い段落|スペースがないため|横スクロールせずに|最後の文はここで/.test(l),
    );
    expect(paragraphLines.length).toBeGreaterThan(1);
  }, TEST_TIMEOUT);

  test("full view wraps long Japanese paragraph", async () => {
    session = await launchCommd({ file: FIXTURE_CJK_WRAP });
    await session.press("f");
    await session.waitForText("f section");
    const text = await session.text({ trimEnd: true });
    expect(text).toContain("最後の文はここで終わります。");
  }, TEST_TIMEOUT);
});
