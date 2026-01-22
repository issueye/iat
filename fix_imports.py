import os

mapping = {
    '"iat/internal/model"': '"iat/common/model"',
    '"iat/internal/pkg/common"': '"iat/common/pkg/common"',
    '"iat/internal/pkg/logger"': '"iat/common/pkg/logger"',
    '"iat/internal/pkg/ai"': '"iat/engine/pkg/ai"',
    '"iat/internal/pkg/tools"': '"iat/engine/pkg/tools"',
    '"iat/internal/pkg/script"': '"iat/engine/pkg/script"',
    '"iat/internal/pkg/indexdb"': '"iat/engine/pkg/indexdb"',
    '"iat/internal/pkg/sse"': '"iat/engine/pkg/sse"',
    '"iat/internal/repo"': '"iat/engine/internal/repo"',
    '"iat/internal/service"': '"iat/engine/internal/service"',
}

root_dir = "iat_new"

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
