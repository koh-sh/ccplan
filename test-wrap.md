# Plan: Word Wrap Verification

This is a preamble with a long sentence to verify that prose text wraps correctly at the pane boundary. If this line appears on a single line without wrapping, the fix did not work. Let's add more text to make absolutely sure this exceeds any reasonable terminal width for testing purposes.

## Step 1: Prose Wrapping

This step contains a long paragraph to test word wrapping behavior. The quick brown fox jumps over the lazy dog. The quick brown fox jumps over the lazy dog again. And once more, the quick brown fox jumps over the lazy dog to ensure this paragraph is long enough to trigger wrapping in most terminal widths.

- This is a bullet point with a long description that should also wrap properly when it exceeds the available width of the detail pane on the right side.
- Short bullet.
- Another long bullet point to verify that list items with extended text content are handled correctly by the glamour word wrap implementation.

## Step 2: Code Block Preservation

Code blocks should NOT be wrapped. They should remain on a single line and be horizontally scrollable:

```go
func thisIsAVeryLongFunctionName(parameterOne string, parameterTwo int, parameterThree bool) (string, error) { return fmt.Sprintf("result: %s %d %v", parameterOne, parameterTwo, parameterThree), nil }
```

Text after code block should wrap normally. This sentence is intentionally long to verify that prose following a code block still wraps as expected at the correct width.

## Step 3: Mixed Content

Here is a mix of prose and code:

First, configure the database connection string in your environment variables. Make sure to use the correct format for your database driver and include all necessary parameters.

```bash
export DATABASE_URL="postgresql://user:password@localhost:5432/mydb?sslmode=disable&connect_timeout=10&application_name=myapp"
```

Then run the migration tool to set up the schema. This will create all necessary tables and indexes for the application to function correctly.

## Step 4: Japanese Text

日本語のテキストも正しく折り返されることを確認します。これは長い日本語の文章で、ペインの幅を超える場合に適切に折り返されるかどうかをテストするためのものです。日本語は全角文字なので、半角文字の約2倍の幅を取ります。

```
これはコードブロック内の日本語です。折り返されずにそのまま表示されるべきです。長い行もそのまま保持されます。コードブロック内ではhorizontal scrollで閲覧します。
```

コードブロックの後の日本語テキストも正しく折り返されるはずです。この文もテスト用に十分な長さにしています。

## Step 5: Mixed CJK and English

日本語と English が混在するテキストのテストです。この文には English words が含まれており、折り返し時に英単語が途中で切られないことを確認します。

The implementation uses softWrapLine first, then falls back to hardWrapCJK only when needed. これにより混在テキストでも英単語の境界が尊重されます。

- リスト内の混在テキスト: This is a test to verify that English words within Japanese text are not broken mid-word during wrapping.
- インデント付き混在: When the line has leading whitespace, both softWrapLine and hardWrapCJK should preserve the indentation on continuation lines correctly.

## Step 6: Indented CJK Text

インデント付きのCJKテキストが正しく折り返されることを確認します:

  インデント付きの日本語テキストです。先頭の空白が継続行にも保持されるかどうかを確認します。これは長い文章なので折り返しが発生するはずです。

    さらに深いインデントのテキストです。このテキストも折り返し時にインデントが維持されることを確認するためのものです。正しく動作すれば各行の先頭に同じ空白が入ります。
