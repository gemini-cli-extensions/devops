import csv
import hashlib
import json
import os
import yaml

# Get directory of the current script
script_dir = os.path.dirname(os.path.abspath(__file__))
json_path = os.path.join(script_dir, 'gcp_cicd_deploy_prompts.json')
output_path = os.path.join(script_dir, 'eval.skill-activation.yaml')

def get_hash(text):
    return hashlib.md5(text.encode('utf-8')).hexdigest()

def generate_yaml():
    # Read JSON
    with open(json_path, 'r') as f:
        prompts_data = json.load(f)

    # Define the base structure
    eval_data = {
        'version': '1',
        'defaults': {
            'agent': 'gemini',
            'provider': 'local',
            'trials': 5,
            'timeout': 60,
            'threshold': 0.8,
            'grader_model': 'gemini-3-flash-preview',
            'docker': {
                'base': 'cicd-evals:latest'
            },
            'env': {
                'GOOGLE_APPLICATION_CREDENTIALS': '~/.config/gcloud/application_default_credentials.json'
            },
            'workspace': [
                {'src': 'scripts/BASE_GEMINI.md', 'dest': '$HOME/.gemini/GEMINI.md'}
            ],
            'environment': {
                'mounts': [
                    '~/.config/gcloud/application_default_credentials.json:/tmp/keys/adc.json:ro'
                ]
            }
        },
        'tasks': []
    }

    skill = 'google-cicd-deploy'

    for item in prompts_data:
        length = item['length']
        specificity = item['specificity']
        prompt = item['prompt']
        prompt_hash = get_hash(prompt)

        task = {
            'name': f'SAT_{skill}_length-{length}_specificity-{specificity}-{prompt_hash}',
            'instruction': prompt,
            'trialConfig': {
                'setup': f'git clone https://github.com/CoasterJX/tinyjam.git /usr/local/google/home/jianxiwang/Desktop/gemini-devops/cicd/tests/workspace/{prompt_hash}',
                'cleanup': f'rm -rf /usr/local/google/home/jianxiwang/Desktop/gemini-devops/cicd/tests/workspace/{prompt_hash}'
            },
            'agentWorkingDir': f'/usr/local/google/home/jianxiwang/Desktop/gemini-devops/cicd/tests/workspace/{prompt_hash}',
            'graders': [
                {
                    'type': 'llm_rubric',
                    'outcome_assertions': [
                        f"Is the exact skill '{skill}' being activated?"
                    ],
                    'weight': 1
                }
            ]
        }
        eval_data['tasks'].append(task)

    # Write YAML
    with open(output_path, 'w') as f:
        yaml.dump(eval_data, f, default_flow_style=False, sort_keys=False, indent=2)

    print(f"Generated {output_path} with {len(eval_data['tasks'])} tasks.")

if __name__ == '__main__':
    generate_yaml()
