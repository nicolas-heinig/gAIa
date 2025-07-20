# gAIa

AI-powered question answering system for my LARP group using local language models.

## Models
- **Embeddings**: `jina/jina-embeddings-v2-base-de`
- **Response Generation**: `cyberwald/llama-3.1-sauerkrautlm-8b-instruct`
- **Categorization**: `nuextract` (currently unused)

## Usage
```sh
# index files in data/
./gAIa index data/

# ask questions
./gAIa ask "Was sind die 5 Paladenstiere?"
```

You can add `-v` to both commands to have more verbose output.

## Data

Each `.txt`, `.docx` and `.pdf` file in the specified directory will be parsed and put into the vector store.

You can add a `CONTEXT.yml` file to categorize all files in the directory:
```yaml
categories:
  - character
files:
  "my_wizard_charcter_sheet.pdf":
    categories:
      - character
      - magic
```
