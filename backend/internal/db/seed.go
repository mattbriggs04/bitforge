package db

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
)

type seedTemplate struct {
	Language    string
	StarterCode string
	Notes       string
}

type seedAsset struct {
	AssetType string
	Name      string
	MIMEType  string
	Content   string
	Hidden    bool
	Metadata  map[string]any
}

type seedCase struct {
	Name          string
	DisplayInput  string
	DisplayExpect string
	Explanation   string
	Payload       map[string]any
	Hidden        bool
	Weight        int
	SortOrder     int
}

type seedProblem struct {
	Slug         string
	Title        string
	Difficulty   string
	Category     string
	ProblemType  string
	Short        string
	Statement    string
	Constraints  string
	Metadata     map[string]any
	Tags         []string
	Templates    []seedTemplate
	Assets       []seedAsset
	JudgeRunner  string
	JudgeConfig  map[string]any
	VisibleCases []seedCase
	HiddenCases  []seedCase
}

func SeedMVP(ctx context.Context, conn *sql.DB) error {
	tx, err := conn.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin seed tx: %w", err)
	}
	defer func() {
		if err != nil {
			_ = tx.Rollback()
		}
	}()

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO users (handle, email)
		VALUES ('demo', 'demo@bitforge.local')
		ON CONFLICT (handle) DO UPDATE SET email = EXCLUDED.email
	`); err != nil {
		return fmt.Errorf("upsert demo user: %w", err)
	}

	problems := defaultSeedProblems()
	for _, problem := range problems {
		problemID, err := upsertProblem(ctx, tx, problem)
		if err != nil {
			return err
		}

		if err := replaceProblemRelations(ctx, tx, problemID, problem); err != nil {
			return err
		}
	}

	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit seed tx: %w", err)
	}
	return nil
}

func upsertProblem(ctx context.Context, tx *sql.Tx, p seedProblem) (string, error) {
	const query = `
		INSERT INTO problems (
			slug,
			title,
			difficulty,
			category,
			problem_type,
			short_description,
			statement_md,
			constraints_md,
			metadata,
			is_published,
			updated_at
		)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9::jsonb, TRUE, now())
		ON CONFLICT (slug)
		DO UPDATE SET
			title = EXCLUDED.title,
			difficulty = EXCLUDED.difficulty,
			category = EXCLUDED.category,
			problem_type = EXCLUDED.problem_type,
			short_description = EXCLUDED.short_description,
			statement_md = EXCLUDED.statement_md,
			constraints_md = EXCLUDED.constraints_md,
			metadata = EXCLUDED.metadata,
			is_published = TRUE,
			updated_at = now()
		RETURNING id
	`
	var id string
	if err := tx.QueryRowContext(
		ctx,
		query,
		p.Slug,
		p.Title,
		p.Difficulty,
		p.Category,
		p.ProblemType,
		p.Short,
		p.Statement,
		p.Constraints,
		mustJSON(p.Metadata),
	).Scan(&id); err != nil {
		return "", fmt.Errorf("upsert problem %s: %w", p.Slug, err)
	}
	return id, nil
}

func replaceProblemRelations(ctx context.Context, tx *sql.Tx, problemID string, p seedProblem) error {
	for _, query := range []string{
		`DELETE FROM problem_tags WHERE problem_id = $1`,
		`DELETE FROM problem_language_templates WHERE problem_id = $1`,
		`DELETE FROM problem_assets WHERE problem_id = $1`,
		`DELETE FROM problem_test_cases WHERE problem_id = $1`,
		`DELETE FROM problem_judge_configs WHERE problem_id = $1`,
	} {
		if _, err := tx.ExecContext(ctx, query, problemID); err != nil {
			return fmt.Errorf("clear problem relations for %s: %w", p.Slug, err)
		}
	}

	for _, tag := range p.Tags {
		if _, err := tx.ExecContext(ctx, `INSERT INTO problem_tags (problem_id, tag) VALUES ($1, $2)`, problemID, tag); err != nil {
			return fmt.Errorf("insert tag %s for %s: %w", tag, p.Slug, err)
		}
	}

	for _, template := range p.Templates {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO problem_language_templates (problem_id, language, starter_code, notes)
			VALUES ($1, $2, $3, $4)
		`, problemID, template.Language, template.StarterCode, template.Notes); err != nil {
			return fmt.Errorf("insert template %s for %s: %w", template.Language, p.Slug, err)
		}
	}

	for _, asset := range p.Assets {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO problem_assets (problem_id, asset_type, name, mime_type, content_text, is_hidden, metadata)
			VALUES ($1, $2, $3, $4, $5, $6, $7::jsonb)
		`, problemID, asset.AssetType, asset.Name, asset.MIMEType, asset.Content, asset.Hidden, mustJSON(asset.Metadata)); err != nil {
			return fmt.Errorf("insert asset %s for %s: %w", asset.Name, p.Slug, err)
		}
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO problem_judge_configs (problem_id, runner, config, updated_at)
		VALUES ($1, $2, $3::jsonb, now())
	`, problemID, p.JudgeRunner, mustJSON(p.JudgeConfig)); err != nil {
		return fmt.Errorf("insert judge config for %s: %w", p.Slug, err)
	}

	insertCase := `
		INSERT INTO problem_test_cases (
			problem_id,
			name,
			display_input,
			display_expected,
			explanation,
			payload,
			is_hidden,
			weight,
			sort_order
		)
		VALUES ($1, $2, $3, $4, $5, $6::jsonb, $7, $8, $9)
	`
	for _, c := range append(p.VisibleCases, p.HiddenCases...) {
		if _, err := tx.ExecContext(
			ctx,
			insertCase,
			problemID,
			c.Name,
			c.DisplayInput,
			c.DisplayExpect,
			c.Explanation,
			mustJSON(c.Payload),
			c.Hidden,
			max(c.Weight, 1),
			c.SortOrder,
		); err != nil {
			return fmt.Errorf("insert test case %s for %s: %w", c.Name, p.Slug, err)
		}
	}

	return nil
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func mustJSON(v any) string {
	if v == nil {
		return "{}"
	}
	bytes, err := json.Marshal(v)
	if err != nil {
		return "{}"
	}
	return string(bytes)
}

func defaultSeedProblems() []seedProblem {
	commonJudgeConfig := map[string]any{
		"c_std":              "c11",
		"compile_timeout_ms": 4000,
		"run_timeout_ms":     2000,
	}

	return []seedProblem{
		{
			Slug:        "bf-strlen",
			Title:       "Reimplement strlen (bf_strlen)",
			Difficulty:  "easy",
			Category:    "Embedded C",
			ProblemType: "libc-reimplementation",
			Short:       "Implement a safe, linear scan for null-terminated strings without calling libc strlen.",
			Statement: `Implement 'size_t bf_strlen(const char *s)'.

Your implementation should count bytes until the first ''\\0''. This exercise looks simple, but interviewers use it to probe pointer walking discipline, off-by-one mistakes, and API edge assumptions.

Rules:
- Do not call 'strlen', 'strnlen', or equivalent helpers.
- Treat input as read-only memory.
- Return the number of bytes before the first null terminator.
`,
			Constraints: `- 's' is a valid pointer to a null-terminated byte sequence.
- Input size is up to 4096 bytes for the visible/hidden tests.
- Time complexity target: 'O(n)'.
- Space complexity target: 'O(1)'.
`,
			Metadata: map[string]any{
				"estimated_minutes": 20,
				"interview_focus":   []string{"pointer arithmetic", "boundary conditions"},
			},
			Tags: []string{"c", "pointers", "libc", "memory"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stddef.h>

size_t bf_strlen(const char *s) {
    (void)s;
    return 0;
}
`,
					Notes: "Focus on pointer traversal and terminating condition.",
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_ascii_word",
					DisplayInput:  "\"firmware\"",
					DisplayExpect: "8",
					Explanation:   "Straight ASCII string length.",
					Payload: map[string]any{
						"code": "const char *input = \"firmware\";\ncase_passed = (bf_strlen(input) == 8);",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_empty",
					DisplayInput:  "\"\"",
					DisplayExpect: "0",
					Explanation:   "Empty string must return zero.",
					Payload: map[string]any{
						"code": "const char *input = \"\";\ncase_passed = (bf_strlen(input) == 0);",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_internal_null",
					Hidden:    true,
					SortOrder: 100,
					Weight:    2,
					Payload:   map[string]any{"code": "const char input[] = {'a', 'b', '\\0', 'x', '\\0'};\ncase_passed = (bf_strlen(input) == 2);"},
				},
				{
					Name:      "hidden_longer_string",
					Hidden:    true,
					SortOrder: 110,
					Weight:    2,
					Payload:   map[string]any{"code": "const char *input = \"interviewprep_for_embedded_systems\";\ncase_passed = (bf_strlen(input) == 34);"},
				},
			},
		},
		{
			Slug:        "bf-memcpy",
			Title:       "Reimplement memcpy (bf_memcpy)",
			Difficulty:  "easy",
			Category:    "Embedded C",
			ProblemType: "libc-reimplementation",
			Short:       "Copy N bytes from source to destination and return destination pointer.",
			Statement: "Implement `void *bf_memcpy(void *dest, const void *src, size_t n)`.\n\n" +
				"For this problem, source and destination are guaranteed not to overlap.\n\n" +
				"Rules:\n" +
				"- Copy exactly `n` bytes.\n" +
				"- Return the original `dest` pointer.\n" +
				"- Do not call `memcpy`, `memmove`, or similar helpers.\n",
			Constraints: "- `dest` and `src` are valid for `n` bytes.\n" +
				"- No overlap in this problem variant.\n" +
				"- Time complexity target: `O(n)`.\n",
			Metadata: map[string]any{
				"estimated_minutes": 25,
				"interview_focus":   []string{"pointer casting", "byte-level copy"},
			},
			Tags: []string{"c", "memory", "pointers", "embedded"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stddef.h>

void *bf_memcpy(void *dest, const void *src, size_t n) {
    (void)dest;
    (void)src;
    (void)n;
    return dest;
}
`,
					Notes: "Prefer unsigned-byte pointer arithmetic.",
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_copy_bytes",
					DisplayInput:  "src=[0x10,0x20,0x30,0x40], n=4",
					DisplayExpect: "dest becomes [0x10,0x20,0x30,0x40]",
					Explanation:   "Basic byte copy semantics.",
					Payload: map[string]any{
						"code": "unsigned char src[4] = {0x10, 0x20, 0x30, 0x40};\nunsigned char dst[4] = {0};\nbf_memcpy(dst, src, 4);\ncase_passed = (memcmp(dst, src, 4) == 0);",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_return_dest",
					DisplayInput:  "dest pointer",
					DisplayExpect: "returned pointer equals dest",
					Explanation:   "Function contract requires returning dest.",
					Payload: map[string]any{
						"code": "unsigned char src[2] = {0xAA, 0xBB};\nunsigned char dst[2] = {0};\nvoid *ret = bf_memcpy(dst, src, 2);\ncase_passed = (ret == (void *)dst);",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_zero_length",
					Hidden:    true,
					SortOrder: 100,
					Payload: map[string]any{
						"code": "unsigned char src[3] = {1,2,3};\nunsigned char dst[3] = {9,9,9};\nbf_memcpy(dst, src, 0);\ncase_passed = (dst[0] == 9 && dst[1] == 9 && dst[2] == 9);",
					},
					Weight: 2,
				},
				{
					Name:      "hidden_partial_copy",
					Hidden:    true,
					SortOrder: 110,
					Payload: map[string]any{
						"code": "unsigned char src[6] = {1,2,3,4,5,6};\nunsigned char dst[6] = {0};\nbf_memcpy(dst, src, 3);\ncase_passed = (dst[0]==1 && dst[1]==2 && dst[2]==3 && dst[3]==0 && dst[4]==0 && dst[5]==0);",
					},
					Weight: 2,
				},
			},
		},
		{
			Slug:        "bf-memmove",
			Title:       "Reimplement memmove (bf_memmove)",
			Difficulty:  "medium",
			Category:    "Embedded C",
			ProblemType: "libc-reimplementation",
			Short:       "Copy bytes safely even when source and destination regions overlap.",
			Statement: "Implement `void *bf_memmove(void *dest, const void *src, size_t n)`.\n\n" +
				"Unlike `memcpy`, memory ranges may overlap. Your logic must choose direction of copy to avoid clobbering unread bytes.\n\n" +
				"Rules:\n" +
				"- Copy exactly `n` bytes.\n" +
				"- Return original `dest`.\n" +
				"- Do not call `memmove` or equivalent helpers.\n",
			Constraints: "- Buffers are valid for `n` bytes.\n" +
				"- Overlap can be forward or backward.\n" +
				"- Time complexity target: `O(n)`.\n",
			Metadata: map[string]any{
				"estimated_minutes": 35,
				"interview_focus":   []string{"overlap handling", "directional copy", "memory safety"},
			},
			Tags: []string{"c", "memory", "overlap", "systems"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stddef.h>

void *bf_memmove(void *dest, const void *src, size_t n) {
    (void)dest;
    (void)src;
    (void)n;
    return dest;
}
`,
					Notes: "Direction matters when ranges overlap.",
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_non_overlap",
					DisplayInput:  "dst[8], src[8], n=8",
					DisplayExpect: "exact byte copy",
					Explanation:   "Should behave like memcpy when no overlap exists.",
					Payload: map[string]any{
						"code": "unsigned char src[4] = {7,8,9,10};\nunsigned char dst[4] = {0};\nbf_memmove(dst, src, 4);\ncase_passed = (memcmp(dst, src, 4) == 0);",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_overlap_forward",
					DisplayInput:  "buffer shifted right by 2",
					DisplayExpect: "source bytes preserved while copying",
					Explanation:   "Classic overlap where dest starts inside source.",
					Payload: map[string]any{
						"code": "unsigned char buf[8] = {1,2,3,4,5,6,7,8};\nbf_memmove(buf + 2, buf, 6);\nunsigned char want[8] = {1,2,1,2,3,4,5,6};\ncase_passed = (memcmp(buf, want, 8) == 0);",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_overlap_backward",
					Hidden:    true,
					SortOrder: 100,
					Weight:    2,
					Payload: map[string]any{
						"code": "unsigned char buf[8] = {1,2,3,4,5,6,7,8};\nbf_memmove(buf, buf + 2, 6);\nunsigned char want[8] = {3,4,5,6,7,8,7,8};\ncase_passed = (memcmp(buf, want, 8) == 0);",
					},
				},
				{
					Name:      "hidden_return_pointer",
					Hidden:    true,
					SortOrder: 110,
					Weight:    1,
					Payload: map[string]any{
						"code": "unsigned char buf[4] = {0};\nvoid *ret = bf_memmove(buf, buf, 4);\ncase_passed = (ret == (void *)buf);",
					},
				},
			},
		},
		{
			Slug:        "ring-buffer-int",
			Title:       "Fixed-Size Ring Buffer (int)",
			Difficulty:  "medium",
			Category:    "Debugging",
			ProblemType: "data-structure-systems",
			Short:       "Implement push/pop semantics for a lock-free-style ring buffer API without dynamic allocation.",
			Statement: `Implement the following API for a fixed-size ring buffer over 'int':

- 'void rb_init(ring_buffer_t *rb, int *storage, size_t capacity)'
- 'bool rb_push(ring_buffer_t *rb, int value)'
- 'bool rb_pop(ring_buffer_t *rb, int *out)'

Behavior:
- 'rb_push' returns 'false' if full.
- 'rb_pop' returns 'false' if empty.
- FIFO order must be preserved.
- Head/tail wrap-around must be correct.
`,
			Constraints: `- No heap allocation.
- Capacity is > 0 for all tests.
- Target operations are 'O(1)'.
`,
			Metadata: map[string]any{
				"estimated_minutes": 40,
				"interview_focus":   []string{"state machine", "modulo arithmetic", "invariants"},
			},
			Tags: []string{"embedded", "ring-buffer", "state-machine", "debugging"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stddef.h>
#include <stdbool.h>

typedef struct {
    int *buf;
    size_t capacity;
    size_t head;
    size_t tail;
    size_t size;
} ring_buffer_t;

void rb_init(ring_buffer_t *rb, int *storage, size_t capacity) {
    (void)rb;
    (void)storage;
    (void)capacity;
}

bool rb_push(ring_buffer_t *rb, int value) {
    (void)rb;
    (void)value;
    return false;
}

bool rb_pop(ring_buffer_t *rb, int *out) {
    (void)rb;
    (void)out;
    return false;
}
`,
					Notes: "Maintain invariants for head, tail, and size across wrap-around.",
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_push_pop",
					DisplayInput:  "capacity=4, push [11,22], pop twice",
					DisplayExpect: "pop order [11,22]",
					Explanation:   "Validates FIFO behavior.",
					Payload: map[string]any{
						"code": "int storage[4] = {0};\nring_buffer_t rb;\nint out = 0;\nrb_init(&rb, storage, 4);\nint ok = rb_push(&rb, 11) && rb_push(&rb, 22);\nok = ok && rb_pop(&rb, &out) && out == 11;\nok = ok && rb_pop(&rb, &out) && out == 22;\ncase_passed = ok;",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_full_rejects_push",
					DisplayInput:  "capacity=2, push [1,2,3]",
					DisplayExpect: "third push returns false",
					Explanation:   "No overwrite behavior for full queue.",
					Payload: map[string]any{
						"code": "int storage[2] = {0};\nring_buffer_t rb;\nrb_init(&rb, storage, 2);\nint ok = rb_push(&rb, 1) && rb_push(&rb, 2);\nok = ok && !rb_push(&rb, 3);\ncase_passed = ok;",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_wraparound",
					Hidden:    true,
					SortOrder: 100,
					Weight:    2,
					Payload: map[string]any{
						"code": "int storage[3] = {0};\nring_buffer_t rb;\nint out = 0;\nrb_init(&rb, storage, 3);\nint ok = rb_push(&rb, 1) && rb_push(&rb, 2) && rb_push(&rb, 3);\nok = ok && rb_pop(&rb, &out) && out == 1;\nok = ok && rb_push(&rb, 4);\nok = ok && rb_pop(&rb, &out) && out == 2;\nok = ok && rb_pop(&rb, &out) && out == 3;\nok = ok && rb_pop(&rb, &out) && out == 4;\ncase_passed = ok;",
					},
				},
				{
					Name:      "hidden_empty_rejects_pop",
					Hidden:    true,
					SortOrder: 110,
					Weight:    1,
					Payload: map[string]any{
						"code": "int storage[2] = {0};\nring_buffer_t rb;\nint out = 99;\nrb_init(&rb, storage, 2);\ncase_passed = !rb_pop(&rb, &out);",
					},
				},
			},
		},
		{
			Slug:        "parse-ipv4-header",
			Title:       "Parse IPv4 Header Fields",
			Difficulty:  "hard",
			Category:    "Networking",
			ProblemType: "packet-parsing",
			Short:       "Parse core IPv4 header metadata from a raw packet buffer with bounds checks.",
			Statement: `Implement:

'int parse_ipv4_header(const unsigned char *packet, size_t len, ipv4_header_t *out)'

Where:
- Return '0' on success.
- Return '-1' on invalid length or unsupported header shape.

The parser should decode:
- version
- ihl (32-bit words)
- total_length
- ttl
- protocol
- src_addr
- dst_addr

Interpret 16/32-bit fields from network byte order to host-order integers.
`,
			Constraints: `- Minimum IPv4 header is 20 bytes.
- Only support version 4.
- Reject packets where 'ihl < 5'.
- Reject packets where provided 'len' is smaller than header length ('ihl * 4').
`,
			Metadata: map[string]any{
				"estimated_minutes": 50,
				"interview_focus":   []string{"binary parsing", "endianness", "bounds checking"},
			},
			Tags: []string{"networking", "packet-parsing", "security", "c"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stddef.h>
#include <stdint.h>

typedef struct {
    uint8_t version;
    uint8_t ihl;
    uint8_t ttl;
    uint8_t protocol;
    uint16_t total_length;
    uint32_t src_addr;
    uint32_t dst_addr;
} ipv4_header_t;

int parse_ipv4_header(const unsigned char *packet, size_t len, ipv4_header_t *out) {
    (void)packet;
    (void)len;
    (void)out;
    return -1;
}
`,
					Notes: "Do strict length checks before reading fields.",
				},
			},
			Assets: []seedAsset{
				{
					AssetType: "reference",
					Name:      "ipv4-header-layout",
					MIMEType:  "text/plain",
					Content: `Byte 0: version (high nibble) + IHL (low nibble)
Bytes 2-3: total length (big-endian)
Byte 8: TTL
Byte 9: Protocol
Bytes 12-15: Source IPv4 address
Bytes 16-19: Destination IPv4 address`,
					Metadata: map[string]any{"kind": "diagram-text"},
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_parse_valid_header",
					DisplayInput:  "20-byte IPv4 header, TTL=64, protocol=6",
					DisplayExpect: "success + expected decoded fields",
					Explanation:   "Basic parse path.",
					Payload: map[string]any{
						"code": "unsigned char pkt[20] = {0x45,0x00,0x00,0x3c,0x00,0x01,0x40,0x00,0x40,0x06,0x00,0x00,192,168,1,10,8,8,8,8};\nipv4_header_t out = {0};\nint rc = parse_ipv4_header(pkt, sizeof(pkt), &out);\ncase_passed = (rc == 0 && out.version == 4 && out.ihl == 5 && out.total_length == 60 && out.ttl == 64 && out.protocol == 6 && out.src_addr == 0xC0A8010A && out.dst_addr == 0x08080808);",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_reject_short_buffer",
					DisplayInput:  "len < 20",
					DisplayExpect: "-1",
					Explanation:   "Parser must reject undersized packet.",
					Payload: map[string]any{
						"code": "unsigned char pkt[10] = {0};\nipv4_header_t out = {0};\ncase_passed = (parse_ipv4_header(pkt, sizeof(pkt), &out) == -1);",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_reject_wrong_version",
					Hidden:    true,
					SortOrder: 100,
					Weight:    2,
					Payload: map[string]any{
						"code": "unsigned char pkt[20] = {0x65,0,0,20,0,0,0,0,32,17,0,0,1,2,3,4,5,6,7,8};\nipv4_header_t out = {0};\ncase_passed = (parse_ipv4_header(pkt, sizeof(pkt), &out) == -1);",
					},
				},
				{
					Name:      "hidden_reject_ihl_too_small",
					Hidden:    true,
					SortOrder: 110,
					Weight:    2,
					Payload: map[string]any{
						"code": "unsigned char pkt[20] = {0x44,0,0,20,0,0,0,0,32,17,0,0,1,2,3,4,5,6,7,8};\nipv4_header_t out = {0};\ncase_passed = (parse_ipv4_header(pkt, sizeof(pkt), &out) == -1);",
					},
				},
			},
		},
		{
			Slug:        "debounce-button-isr",
			Title:       "Debounce Button with Timer ISR Simulation",
			Difficulty:  "hard",
			Category:    "Embedded C",
			ProblemType: "hardware-simulation",
			Short:       "Implement a shift-register debouncer polled from a simulated timer ISR using a GPIO input register.",
			Statement: `Implement a timer-driven software debouncer for a simulated button input.

You are given a GPIO input register snapshot on every timer tick. Your code should:
- track a selected pin
- shift in one sampled bit per tick
- detect stable transitions only when the 8-bit history is all ones or all zeros
- latch edge events and expose consume APIs

Implement:
- 'void debounce_init(debounce_t *db, uint8_t pin)'
- 'void debounce_timer_isr(debounce_t *db, uint32_t gpio_in_reg)'
- 'int debounce_consume_rising_edge(debounce_t *db)'
- 'int debounce_consume_falling_edge(debounce_t *db)'

Simulation model:
- Timer ISR calls 'debounce_timer_isr' each tick.
- Sampled bit is '(gpio_in_reg >> db->pin) & 1'.
- Update history with a single-byte shift register:
  'history = (history << 1) | sample'
- If history becomes 0xFF and previous stable state was low, latch one rising edge.
- If history becomes 0x00 and previous stable state was high, latch one falling edge.
- Consume functions return 1 once per latched edge, then clear the latch.
`,
			Constraints: `- Use only byte-sized history for debouncing.
- Do not use dynamic allocation.
- Pin index in tests is between 0 and 15.
- Debounce threshold is exactly 8 consistent samples.
- This problem simulates ISR polling; do not block or busy-wait.
`,
			Metadata: map[string]any{
				"estimated_minutes": 55,
				"interview_focus":   []string{"interrupt-driven design", "register access", "edge detection", "state machines"},
				"simulated_hardware": map[string]any{
					"gpio_register":  "gpio_in_reg",
					"sampling_model": "timer_isr_poll",
					"debounce_bits":  8,
				},
			},
			Tags: []string{"embedded", "firmware", "interrupts", "debounce", "registers"},
			Templates: []seedTemplate{
				{
					Language: "c",
					StarterCode: `#include <stdint.h>

typedef struct {
    uint8_t pin;
    uint8_t history;
    uint8_t stable_state;
    uint8_t rose_event;
    uint8_t fell_event;
} debounce_t;

void debounce_init(debounce_t *db, uint8_t pin) {
    (void)db;
    (void)pin;
}

void debounce_timer_isr(debounce_t *db, uint32_t gpio_in_reg) {
    (void)db;
    (void)gpio_in_reg;
}

int debounce_consume_rising_edge(debounce_t *db) {
    (void)db;
    return 0;
}

int debounce_consume_falling_edge(debounce_t *db) {
    (void)db;
    return 0;
}
`,
					Notes: "Treat the timer callback as ISR-safe logic: update state, latch events, return quickly.",
				},
			},
			Assets: []seedAsset{
				{
					AssetType: "reference",
					Name:      "debounce-register-model",
					MIMEType:  "text/plain",
					Content: `Each timer tick provides one 32-bit GPIO input snapshot.
To sample button pin N:
  sample = (gpio_in_reg >> N) & 1

8-bit debounce history:
  history = (history << 1) | sample

history == 0xFF => stable high
history == 0x00 => stable low`,
					Metadata: map[string]any{"kind": "simulation-notes"},
				},
			},
			JudgeRunner: "c_assert_harness_v1",
			JudgeConfig: commonJudgeConfig,
			VisibleCases: []seedCase{
				{
					Name:          "sample_latches_rising_after_stable_high",
					DisplayInput:  "Bouncy low-to-high sequence on pin 5",
					DisplayExpect: "Exactly one rising edge event",
					Explanation:   "Rising edge should trigger only after 8 consecutive high samples.",
					Payload: map[string]any{
						"code": "debounce_t db;\nconst uint32_t B = (1u << 5);\ndebounce_init(&db, 5);\nuint32_t seq[] = {0,0,0,0,B,0,B,0,B,B,B,B,B,B,B,B};\nfor (size_t i = 0; i < sizeof(seq)/sizeof(seq[0]); i++) {\n    debounce_timer_isr(&db, seq[i]);\n}\nint first = debounce_consume_rising_edge(&db);\nint second = debounce_consume_rising_edge(&db);\ncase_passed = (first == 1 && second == 0 && debounce_consume_falling_edge(&db) == 0);",
					},
					SortOrder: 1,
				},
				{
					Name:          "sample_latches_falling_after_stable_low",
					DisplayInput:  "High-to-low bounce sequence on pin 2",
					DisplayExpect: "Exactly one falling edge event",
					Explanation:   "Falling edge must wait for 8 consecutive low samples.",
					Payload: map[string]any{
						"code": "debounce_t db;\nconst uint32_t B = (1u << 2);\ndebounce_init(&db, 2);\nuint32_t seq[] = {B,B,B,B,B,B,B,B,B,0,B,0,0,0,0,0,0,0,0,0};\nfor (size_t i = 0; i < sizeof(seq)/sizeof(seq[0]); i++) {\n    debounce_timer_isr(&db, seq[i]);\n}\nint first = debounce_consume_falling_edge(&db);\nint second = debounce_consume_falling_edge(&db);\ncase_passed = (first == 1 && second == 0 && debounce_consume_rising_edge(&db) == 0);",
					},
					SortOrder: 2,
				},
			},
			HiddenCases: []seedCase{
				{
					Name:      "hidden_ignores_short_chatter",
					Hidden:    true,
					SortOrder: 100,
					Weight:    2,
					Payload: map[string]any{
						"code": "debounce_t db;\nconst uint32_t B = (1u << 7);\ndebounce_init(&db, 7);\nuint32_t seq[] = {0,B,0,B,0,B,0,B,0,B,0,0,0};\nfor (size_t i = 0; i < sizeof(seq)/sizeof(seq[0]); i++) {\n    debounce_timer_isr(&db, seq[i]);\n}\ncase_passed = (debounce_consume_rising_edge(&db) == 0 && debounce_consume_falling_edge(&db) == 0);",
					},
				},
				{
					Name:      "hidden_samples_correct_pin_only",
					Hidden:    true,
					SortOrder: 110,
					Weight:    2,
					Payload: map[string]any{
						"code": "debounce_t db;\nconst uint32_t P = (1u << 11);\ndebounce_init(&db, 11);\nuint32_t seq[] = {\n  P, P|1u, P|2u, P|4u, P|8u, P|16u, P|32u, P|64u,\n  0u, 1u, 2u, 4u, 8u, 16u, 32u, 64u\n};\nfor (size_t i = 0; i < sizeof(seq)/sizeof(seq[0]); i++) {\n    debounce_timer_isr(&db, seq[i]);\n}\nint rise = debounce_consume_rising_edge(&db);\nint fall = debounce_consume_falling_edge(&db);\ncase_passed = (rise == 1 && fall == 1);",
					},
				},
				{
					Name:      "hidden_rising_not_retriggered_without_new_transition",
					Hidden:    true,
					SortOrder: 120,
					Weight:    1,
					Payload: map[string]any{
						"code": "debounce_t db;\nconst uint32_t B = (1u << 1);\ndebounce_init(&db, 1);\nuint32_t seq[] = {0,0,0,0,0,0,0,0,B,B,B,B,B,B,B,B,B,B,B,B};\nfor (size_t i = 0; i < sizeof(seq)/sizeof(seq[0]); i++) {\n    debounce_timer_isr(&db, seq[i]);\n}\nint first = debounce_consume_rising_edge(&db);\nint second = debounce_consume_rising_edge(&db);\ncase_passed = (first == 1 && second == 0 && debounce_consume_falling_edge(&db) == 0);",
					},
				},
			},
		},
	}
}
