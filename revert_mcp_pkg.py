import os

# Revert: new -> old
target_pkg = "github.com/mark3labs/mcp-go"
wrong_pkg = "github.com/mark3labs/mcphost"

root_dir = "iat_new"

for root, dirs, files in os.walk(root_dir):
    for file in files:
        if file.endswith(".go") or file == "go.mod":
            path = os.path.join(root, file)
            with open(path, "r", encoding="utf-8") as f:
                content = f.read()
            
            if wrong_pkg in content:
                print(f"Reverting {path}")
                new_content = content.replace(wrong_pkg, target_pkg)
                with open(path, "w", encoding="utf-8") as f:
                    f.write(new_content)
