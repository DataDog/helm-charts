"""Render contract tests. Run: python3 -m unittest discover -s test/synthetics-private-location"""
import copy
import json
from pathlib import Path
import subprocess
import tempfile
import unittest

import yaml

CHART = Path(__file__).resolve().parents[2] / 'charts/synthetics-private-location'
BASE = {
    'image': {'tag': '1.77.0'},
    'multiContainer': {
        'enabled': True,
        'browserPool': {'adminTokenSecret': {'name': 'pool-token', 'key': 'admin'}},
        'networkPolicy': {'mode': 'external'},
    },
}


def render(values):
    with tempfile.NamedTemporaryFile(mode='w', suffix='.json') as stream:
        json.dump(values, stream)
        stream.flush()
        return subprocess.run(['helm', 'template', 'test', str(CHART), '-f', stream.name],
                              capture_output=True, text=True)


class ChartTest(unittest.TestCase):
    def setUp(self):
        self.values = copy.deepcopy(BASE)

    def docs(self):
        result = render(self.values)
        self.assertEqual(result.returncode, 0, result.stderr)
        return {d['kind']: d for d in yaml.safe_load_all(result.stdout) if d}

    def pod(self):
        return self.docs()['Deployment']['spec']['template']['spec']

    def fails(self, message):
        result = render(self.values)
        self.assertNotEqual(result.returncode, 0)
        self.assertIn(message, result.stderr)

    def managed(self):
        policy = self.values['multiContainer']['networkPolicy']
        policy.update(mode='managed', dns={'to': [{'ipBlock': {'cidr': '169.254.20.10/32'}}]})
        return policy

    def test_legacy_remains_usable_without_new_settings(self):
        self.values = {}
        docs = self.docs()
        self.assertNotIn('NetworkPolicy', docs)
        self.assertNotIn('SecurityContextConstraints', docs)
        pod = docs['Deployment']['spec']['template']['spec']
        self.assertEqual(len(pod['containers']), 1)
        self.assertNotIn('command', pod['containers'][0])

    def test_secret_reference_and_hardening(self):
        docs = self.docs()
        self.assertNotIn('Secret', docs)
        self.assertNotIn('NetworkPolicy', docs)
        pod = docs['Deployment']['spec']['template']['spec']
        self.assertFalse(pod['automountServiceAccountToken'])
        containers = {c['name']: c for c in pod['containers']}
        self.assertEqual(len(containers), 4)
        for name, container in containers.items():
            sc = container['securityContext']
            self.assertFalse(sc['allowPrivilegeEscalation'])
            self.assertTrue(sc['readOnlyRootFilesystem'])
            self.assertTrue(sc['runAsNonRoot'])
            self.assertEqual(sc['capabilities']['drop'], ['ALL'])
            tokens = [v for v in container.get('env', []) if v['name'] == 'BROWSER_POOL_ADMIN_TOKEN']
            if name in ('synthetics-private-location', 'browser-pool'):
                self.assertEqual(tokens, [{'name': 'BROWSER_POOL_ADMIN_TOKEN', 'valueFrom': {
                    'secretKeyRef': {'name': 'pool-token', 'key': 'admin'}}}])
            else:
                self.assertEqual(tokens, [])
        self.assertEqual(containers['deno-executor']['resources']['requests']['memory'], '1Gi')
        self.assertEqual(containers['deno-executor']['resources']['limits']['memory'], '1Gi')
        self.assertIn("-name 'browser-*' -user nobody", containers['browser-pool']['command'][-1])
        self.assertIn('exec node', containers['browser-pool']['command'][-1])

    def test_required_settings(self):
        for field in ('name', 'key'):
            with self.subTest(field=field):
                self.values = copy.deepcopy(BASE)
                self.values['multiContainer']['browserPool']['adminTokenSecret'][field] = ''
                self.fails('adminTokenSecret.' + field)
        for mode in ('', 'unknown'):
            self.values = copy.deepcopy(BASE)
            self.values['multiContainer']['networkPolicy']['mode'] = mode
            self.fails('explicitly set')
        self.values = copy.deepcopy(BASE)
        self.values['multiContainer']['networkPolicy']['enabled'] = True
        self.fails('has been replaced')
        self.values = copy.deepcopy(BASE)
        self.values['env'] = [{'name': 'BROWSER_POOL_ADMIN_TOKEN', 'value': 'unsafe'}]
        self.fails('not env')

    def test_image_versions(self):
        for tag in ('1.74.0', '1.76.9', '1.77.0-rc.1', 'latest', ''):
            with self.subTest(tag=tag):
                self.values['image']['tag'] = tag
                self.fails('>= 1.77.0')
        for tag in ('1.77.0', '1.78.0', '2.0.0'):
            self.values['image']['tag'] = tag
            self.docs()

    def test_dns_is_required_and_scoped(self):
        policy = self.managed()
        policy['dns']['to'] = []
        self.fails('must specify DNS')
        policy['dns']['to'] = [{}]
        self.fails('each multiContainer.networkPolicy.dns.to entry')
        peer = {'namespaceSelector': {'matchLabels': {'kubernetes.io/metadata.name': 'kube-system'}},
                'podSelector': {'matchLabels': {'k8s-app': 'kube-dns'}}}
        policy['dns']['to'] = [peer]
        rules = self.docs()['NetworkPolicy']['spec']['egress']
        self.assertEqual(rules[0], {'to': [peer], 'ports': [
            {'port': 53, 'protocol': 'UDP'}, {'port': 53, 'protocol': 'TCP'}]})
        self.assertEqual(len(rules), 2)
        blocks = rules[1]['to']
        self.assertIn('10.0.0.0/8', blocks[0]['ipBlock']['except'])
        self.assertIn('fc00::/7', blocks[1]['ipBlock']['except'])

    def test_full_family_blocks_and_exceptions(self):
        policy = self.managed()
        policy.update(blockedIPRanges={'IPv4': ['0.0.0.0/0'], 'IPv6': ['::/0']},
                      allowedIPRanges={'IPv4': ['10.0.0.1'], 'IPv6': ['fd00::1']},
                      additionalEgress=[{'to': [{'ipBlock': {'cidr': '10.2.3.4/32'}}],
                                         'ports': [{'protocol': 'TCP', 'port': 3128}]}])
        rules = self.docs()['NetworkPolicy']['spec']['egress']
        self.assertEqual(len(rules), 3)
        self.assertEqual(rules[1], policy['additionalEgress'][0])
        self.assertEqual(rules[2]['to'], [{'ipBlock': {'cidr': '10.0.0.1/32'}},
                                         {'ipBlock': {'cidr': 'fd00::1/128'}}])
        policy['enableDefaultBlockedIpRanges'] = False
        policy['blockedIPRanges'] = {'IPv4': ['192.0.2.1'], 'IPv6': []}
        rules = self.docs()['NetworkPolicy']['spec']['egress']
        self.assertEqual(rules[-1]['to'], [
            {'ipBlock': {'cidr': '0.0.0.0/0', 'except': ['192.0.2.1/32']}},
            {'ipBlock': {'cidr': '::/0'}}])

    def test_sysctl_override_and_duplicates(self):
        self.values['securityContext'] = {'runAsGroup': 1234}
        self.assertEqual(self.pod()['securityContext']['sysctls'], [
            {'name': 'net.ipv4.ping_group_range', 'value': '1234 1234'}])
        settings = [{'name': 'net.ipv4.ping_group_range', 'value': '0 65535'},
                    {'name': 'net.ipv4.ip_local_port_range', 'value': '1024 65535'}]
        self.values['podSecurityContext'] = {'sysctls': settings}
        self.assertEqual(self.pod()['securityContext']['sysctls'], settings)
        settings.append(settings[0])
        self.fails('duplicate name net.ipv4.ping_group_range')

    def test_traceroute_disabled_and_scc(self):
        self.values['multiContainer']['tracerouteServer'] = {'enabled': False}
        self.values['podSecurity'] = {'securityContextConstraints': {'create': True}}
        docs = self.docs()
        self.assertEqual(docs['SecurityContextConstraints']['allowedCapabilities'], [])
        containers = docs['Deployment']['spec']['template']['spec']['containers']
        self.assertEqual(len(containers), 3)
        self.assertFalse(any(e['name'] == 'DATADOG_TRACEROUTE_USE_HTTP' for e in containers[0]['env']))

    def test_config_sources_and_stable_selectors(self):
        reference = self.docs()['Deployment']
        for source, field in [('configSecret', 'secret'), ('configConfigMap', 'configMap')]:
            self.values[source] = 'existing-config'
            doc = self.docs()['Deployment']
            self.assertEqual(doc['spec']['selector'], reference['spec']['selector'])
            volume = next(v for v in doc['spec']['template']['spec']['volumes'] if v['name'] == 'worker-config')
            self.assertIn(field, volume)
            del self.values[source]


if __name__ == '__main__':
    unittest.main()
