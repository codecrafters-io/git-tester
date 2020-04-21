import sys
import os
import zlib

import hashlib
import pathlib


# print(sys.argv)

command = sys.argv[1]
if command == "init":
    os.mkdir(".git")
    os.mkdir(".git/objects")
    os.mkdir(".git/refs")
    with open(".git/HEAD", "w") as f:
        f.write("ref: refs/heads/master\n")

    print("Initialized git directory")
elif command == "hash-object":
    assert sys.argv[2] == "-w"
    filepath = sys.argv[3]
    contents = open(filepath).read()
    header = f"blob {len(contents)}\0"
    store = (header + contents).encode()
    sha = hashlib.sha1(store).hexdigest()
    print(sha)
    zlib_store = zlib.compress(store)
    path = f".git/objects/{sha[0:2]}/{sha[2:]}"
    os.makedirs(os.path.dirname(path), exist_ok=True)
    open(path, "wb").write(zlib_store)
else:
    raise RuntimeError(f"Unknown command: #{command}")
