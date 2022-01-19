# config/puma.rb
port 3000
workers 3
preload_app!

activate_control_app 'tcp://127.0.0.1:9292', { no_token: true }

# Load the metrics plugin
plugin 'metrics'
# Bind the metric server to "url". "tcp://" is the only accepted protocol.
#
# The default is "tcp://0.0.0.0:9393".
metrics_url 'tcp://0.0.0.0:9393'
