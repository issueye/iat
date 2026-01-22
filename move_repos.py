import os
import shutil

src_dir = "iat_new/engine/internal"
dst_dir = "iat_new/engine/internal/repo"

if not os.path.exists(dst_dir):
    os.makedirs(dst_dir)

for file in os.listdir(src_dir):
    if file.endswith("_repo.go"):
        src_path = os.path.join(src_dir, file)
        dst_path = os.path.join(dst_dir, file)
        print(f"Moving {src_path} to {dst_path}")
        shutil.move(src_path, dst_path)
