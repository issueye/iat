import os

mapping = {
    '"iat/internal/pkg/consts"': '"iat/common/pkg/consts"',
    '"iat/internal/pkg/tools/builtin"': '"iat/engine/pkg/tools/builtin"',
    '"iat/internal/pkg/tools/script"': '"iat/engine/pkg/tools/script"',
    '"iat/internal/pkg/db"': '"iat/common/pkg/db"',
}

root_dir = "iat_new/engine"

for root, dirs, files in os.walk(root_dir):
    for file in files:
        if file.endswith(".go"):
            path = os.path.join(root, file)
            with open(path, "r", encoding="utf-8") as f:
                content = f.read()
            
            new_content = content
            for old, new in mapping.items():
                new_content = new_content.replace(old, new)
            
            if new_content != content:
                print(f"Updating {path}")
                with open(path, "w", encoding="utf-8") as f:
                    f.write(new_content)
