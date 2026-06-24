import re

def remove_watchdog_from_rover():
    with open('rover.go', 'r', encoding='utf-8') as f:
        code = f.read()

    # 1. Remove invocation
    invocation_pattern = r'watchdogCtx,\s*watchdogCancel\s*:=\s*context\.WithCancel\(context\.Background\(\)\)\s*(if\s*r\.GetConfig\(\)\.EnableFailover\s*\{.*?\})?'
    code = re.sub(invocation_pattern, r'watchdogCtx, watchdogCancel := context.WithCancel(context.Background())', code, flags=re.DOTALL)
    
    # Actually wait, `watchdogCtx` is used by the healthTicker just below it.
    # Let's just remove the `if r.GetConfig().EnableFailover { go r.runFailoverWatchdog(watchdogCtx) }`
    code = re.sub(
        r'if r\.GetConfig\(\)\.EnableFailover\s*\{\s*go r\.runFailoverWatchdog\(watchdogCtx\)\s*\}',
        '',
        code
    )

    # 2. Remove function definition
    func_pattern = r'func \(r \*Rover\) runFailoverWatchdog\(ctx context\.Context\) \{.*'
    # We will just split and remove the end since it's the last function in the file
    func_idx = code.find('func (r *Rover) runFailoverWatchdog')
    if func_idx != -1:
        code = code[:func_idx]

    with open('rover.go', 'w', encoding='utf-8') as f:
        f.write(code)

def remove_watchdog_from_config():
    with open('config.go', 'r', encoding='utf-8') as f:
        code = f.read()

    code = re.sub(r'NotifyOnFailover\s+bool\s+`yaml:"notify_on_failover"`', '', code)
    code = re.sub(r'EnableFailover\s+bool\s+`yaml:"enable_failover"`', '', code)
    code = re.sub(r'FailoverInterval\s+int\s+`yaml:"failover_interval".*?\n', '', code)
    code = re.sub(r'FailoverMaxFails\s+int\s+`yaml:"failover_max_fails".*?\n', '', code)
    
    # Remove from default YAML template
    code = re.sub(r'# 是否啟用秒級急救機制.*?\n.*?\n', '', code, flags=re.DOTALL)
    code = re.sub(r'# 秒級急救機制的偵測間隔.*?\n.*?\n', '', code, flags=re.DOTALL)
    code = re.sub(r'# 秒級急救機制的容忍失敗次數.*?\n.*?\n', '', code, flags=re.DOTALL)
    code = re.sub(r'# 啟用秒級急救機制通知.*?\n.*?\n', '', code, flags=re.DOTALL)

    # Remove from struct instantiation
    code = re.sub(r'EnableFailover:\s*true,\n', '', code)
    code = re.sub(r'FailoverInterval:\s*3,\n', '', code)
    code = re.sub(r'FailoverMaxFails:\s*3,\n', '', code)

    with open('config.go', 'w', encoding='utf-8') as f:
        f.write(code)

def remove_watchdog_from_yaml():
    try:
        with open('rover_config.yaml', 'r', encoding='utf-8') as f:
            code = f.read()
            
        code = re.sub(r'# 是否啟用秒級急救機制.*?\n.*?\n', '', code, flags=re.DOTALL)
        code = re.sub(r'# 秒級急救機制的偵測間隔.*?\n.*?\n', '', code, flags=re.DOTALL)
        code = re.sub(r'# 秒級急救機制的容忍失敗次數.*?\n.*?\n', '', code, flags=re.DOTALL)
        code = re.sub(r'# 啟用秒級急救機制通知.*?\n.*?\n', '', code, flags=re.DOTALL)
        code = re.sub(r'enable_failover:.*?\n', '', code)
        code = re.sub(r'failover_interval:.*?\n', '', code)
        code = re.sub(r'failover_max_fails:.*?\n', '', code)
        code = re.sub(r'notify_on_failover:.*?\n', '', code)

        with open('rover_config.yaml', 'w', encoding='utf-8') as f:
            f.write(code)
    except FileNotFoundError:
        pass


remove_watchdog_from_rover()
remove_watchdog_from_config()
remove_watchdog_from_yaml()
