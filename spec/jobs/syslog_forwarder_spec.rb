require 'bosh/template/test'
require 'yaml'

describe 'syslog_forwarder' do
  let(:release) { Bosh::Template::Test::ReleaseDir.new(File.join(File.dirname(__FILE__), '../..')) }
  let(:job) { release.job('syslog_forwarder') }

  describe 'bin/blackbox_ctl' do
    let(:template) { job.template('bin/blackbox_ctl') }

    it 'sources privdrop_utils.sh from the blackbox package' do
      rendered = template.render({})
      expect(rendered).to include('source /var/vcap/packages/blackbox/scripts/privdrop_utils.sh')
    end

    it 'runs blackbox using run_as_syslog' do
      rendered = template.render({})
      expect(rendered).to match(/run_as_syslog \/var\/vcap\/packages\/blackbox\/bin\/blackbox \\/)
    end

    it 'does not use chpst' do
      rendered = template.render({})
      expect(rendered).not_to include('chpst')
    end

    it 'grants cap_dac_read_search by default when respect_file_permissions is false' do
      rendered = template.render({})
      expect(rendered).to include('setcap cap_dac_read_search+ep /var/vcap/packages/blackbox/bin/blackbox')
    end

    context 'when syslog.respect_file_permissions is true' do
      let(:rendered) { template.render({ 'syslog' => { 'respect_file_permissions' => true } }) }

      it 'does not grant file capabilities' do
        expect(rendered).not_to include('setcap')
      end
    end

    it 'exports GOMAXPROCS=1 by default when limit_cpu is true' do
      rendered = template.render({})
      expect(rendered).to include('export GOMAXPROCS=1')
    end

    context 'when syslog.blackbox.limit_cpu is false' do
      let(:rendered) { template.render({ 'syslog' => { 'blackbox' => { 'limit_cpu' => false } } }) }

      it 'does not export GOMAXPROCS' do
        expect(rendered).not_to include('export GOMAXPROCS')
      end
    end
  end

  describe 'syslog-release.conf' do
    let(:template) { job.template('config/syslog-release.conf') }

    context 'when syslog.use_tcp_for_file_forwarding_local_transport is false (default)' do
      let(:rendered) { template.render({}) }

      it 'does not configure a TCP listener' do
        expect(rendered).not_to include('imtcp')
        expect(rendered).not_to include('$InputTCPServerRun')
      end

      it 'keeps the localhost UDP listener' do
        expect(rendered).to include('$ModLoad imudp')
        expect(rendered).to include('$UDPServerAddress localhost')
        expect(rendered).to include('$UDPServerRun 514')
      end
    end

    context 'when syslog.use_tcp_for_file_forwarding_local_transport is true' do
      let(:rendered) { template.render({ 'syslog' => { 'use_tcp_for_file_forwarding_local_transport' => true } }) }

      it 'binds the TCP listener to 127.0.0.1 only' do
        expect(rendered).to match(/^input\(type="imtcp" port="514" address="127\.0\.0\.1"\)$/)
        expect(rendered).to include('module(load="imtcp")')
        expect(rendered).not_to include('$ModLoad imtcp')
        expect(rendered).not_to include('$InputTCPServerRun')
      end

      it 'keeps the localhost UDP listener' do
        expect(rendered).to include('$ModLoad imudp')
        expect(rendered).to include('$UDPServerAddress localhost')
        expect(rendered).to include('$UDPServerRun 514')
      end
    end

    context 'when syslog.migration.disabled is true' do
      let(:rendered) do
        template.render({ 'syslog' => { 'migration' => { 'disabled' => true }, 'use_tcp_for_file_forwarding_local_transport' => true } })
      end

      it 'renders nothing' do
        expect(rendered.strip).to be_empty
      end
    end
  end
end
