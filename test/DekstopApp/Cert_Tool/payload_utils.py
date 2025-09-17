def generate_payload_utils():
    import os
    import json
    import subprocess

    def load_sample_data(sample_file):
        if os.path.exists(sample_file):
            with open(sample_file, 'r') as f:
                return json.load(f)
        return {}

    def generate_payload(data):
        # Placeholder for payload generation logic
        return data

    def execute_command(command):
        try:
            result = subprocess.run(command, shell=True, capture_output=True, text=True)
            return result.stdout.strip(), result.returncode
        except Exception as e:
            return str(e), 1

    def save_payload_to_file(payload, filename):
        with open(filename, 'w') as f:
            json.dump(payload, f, indent=4)

    sample_data = load_sample_data('sample.json')
    payload = generate_payload(sample_data)
    save_payload_to_file(payload, 'generated_payload.json')

    return {
        "load_sample_data": load_sample_data,
        "generate_payload": generate_payload,
        "execute_command": execute_command,
        "save_payload_to_file": save_payload_to_file
    }

payload_utils = generate_payload_utils()