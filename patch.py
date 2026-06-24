import sys

def main():
    with open('web.go', 'r', encoding='utf-8') as f:
        content = f.read()

    start_str = 'func handleIndex(w http.ResponseWriter, r *http.Request) {\n'
    idx = content.find(start_str)
    
    if idx == -1:
        print("Could not find handleIndex")
        return

    new_content = content[:idx] + start_str + '\tw.Header().Set("Content-Type", "text/html; charset=utf-8")\n\tw.Write(indexHTML)\n}\n'
    
    with open('web.go', 'w', encoding='utf-8') as f:
        f.write(new_content)
    print("Successfully patched web.go")

if __name__ == '__main__':
    main()
