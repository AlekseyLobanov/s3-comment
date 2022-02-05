wrk.method = "POST"
wrk.body = '{"text": "## Emphasis\\n\\n**This is bold text**\\n\\n__This is bold text__\\n\\n*This is italic text*\\n\\n_This is italic text_\\n\\n~~Strikethrough~~"}'
wrk.headers["Content-Type"] = "application/json"
