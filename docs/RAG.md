# RAG & Memory Implementation Plan

This document outlines the architecture and implementation steps for adding Retrieval-Augmented Generation (RAG) and Long-Term Memory to TuskBot.

## Overview

The goal is to allow the agent to recall facts and context from past conversations without overloading the context window. This involves a hybrid approach:
1.  **Semantic Search**: Finding relevant past messages.
2.  **Knowledge Extraction**: Periodically summarizing conversations into atomic facts.
3.  **Context Assembly**: Injecting these findings into the System Prompt.

## 1. Database Schema & Storage

The database schema is defined in `internal/storage/sqlite/migrations/20260207210202_init.sql`.

### Existing Schema

**Messages Table (`messages`)**
- Stores chat history.
- Columns: `id`, `session_id`, `role`, `content`, `tool_calls`, `tool_call_id`, `embedding` (BLOB), `created_at`, `extracted` (BOOLEAN).
- `extracted` flag tracks which messages have been processed by the background knowledge extractor.

**Vector Table (`messages_vec`)**
- Virtual table using `vec0` for vector search on messages.
- Dimension: 768.

**Knowledge Table (`knowledge`)**
- Stores extracted facts.
- Columns: `id`, `fact`, `category`, `source`, `embedding` (BLOB), `created_at`, `updated_at`, `fact_hash`.

**Vector Table (`knowledge_vec`)**
- Virtual table using `vec0` for vector search on knowledge facts.
- Dimension: 768.

### Repository Interface Updates

New methods needed in `internal/core` or `internal/storage`:


```golang
type KnowledgeRepository interface {


// Save extracted fact

SaveFact(ctx context.Context, fact core.StoredKnowledge) error



// Search knowledge and messages combined (The "Union Query")

SearchContext(ctx context.Context, vector []float32, limitKnowledge, limitHistory int)
([]core.ContextItem, error)



// Mark messages as processed by the extractor

MarkMessagesExtracted(ctx context.Context, messageIDs []int64) error



// Get unextracted messages for the background job

GetUnextractedMessages(ctx context.Context, limit int) ([]core.StoredMessage, error)


}

```

## 2. Tool Output Truncation (Anti-Spaghetti)



To prevent context explosion when tools return massive outputs (e.g., `cat large_file.log`), we
implement a middleware/interceptor in the Agent loop.



**Logic:**

1.  Execute Tool.

2.  Check output length (e.g., > 2000 chars).

3.  If too long, truncate and append a warning.


```golang
const MaxToolOutputLen = 2000

if len(result) > MaxToolOutputLen {


truncated := result[:MaxToolOutputLen]

result = fmt.Sprintf(

    "%s\n... [Output truncated. Original size: %d bytes. Hint: Use standard unix tools like grep/head
to read specific parts]",

    truncated,

    len(rawResult),

)


}
```

## 3. The RAG Context Builder

A new service `ContextBuilder` is responsible for assembling the final prompt sent to the LLM.



**Workflow:**

1.  **User Input**: `input`

2.  **Embed**: `vector = embed(input)`

3.  **Search**:

    *   Top-3 Facts from `knowledge` table (using `vector`).

    *   Top-2 Messages from `messages` table (using `vector`).

4.  **Retrieve Recent History**: Last N messages (standard SQL `ORDER BY created_at DESC`).

5.  **Assemble**:


[System Prompt (from disk)]

Relevant Facts (Global Context):

• Fact 1

• Fact 2

Related Conversation History (Semantic Context):

• User: ...

• Assistant: ...

Current Conversation (Recent History):

• User: ...

• Assistant: ...




## 4. Background Knowledge Extraction



A background routine (Cron/Ticker) that converts raw conversation logs into atomic facts.



**Algorithm:**

1.  **Select**: `SELECT * FROM messages WHERE role != 'system' AND extracted = FALSE LIMIT 20`.

2.  **LLM Call**: "Extract atomic facts from this conversation...".

3.  **Embed**: Generate vectors for the extracted facts.

4.  **Save**: Insert into `knowledge` and `knowledge_vec`.

5.  **Update**: Set `extracted = TRUE` for processed messages.



## 5. Implementation Steps



1.  **SQL**: (Done) Migrations for `embedding` columns and `knowledge` table are present in
    `internal/storage/sqlite/migrations`.

2.  **Storage**: Implement the `SearchContext` (Union Query) and `SaveFact` in
    `internal/storage/sqlite`.

3.  **Agent**: Add the Tool Output Truncation logic.

4.  **Agent**: Integrate Embedding generation on user input.

5.  **Service**: Implement the `ContextBuilder` to merge System, Knowledge, and History.

6.  **Background**: Implement the Extraction Routine.