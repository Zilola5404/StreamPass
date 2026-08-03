import 'package:flutter/material.dart';

import '../services/auth_service.dart';
import '../services/streampass_api.dart';
import '../main.dart' show navigateToLogin;
import '../theme/app_theme.dart';

/// Read-only list of relay servers. Reuses the same GET /servers call
/// already made on Home — no new backend surface needed. Selection UI
/// (choosing a specific relay instead of "best available") is deliberately
/// not built yet: with a single live relay in production right now, a
/// picker has no real functionality to test.
class ServersScreen extends StatefulWidget {
  final StreamPassApi api;
  final AuthService? authService;
  const ServersScreen({super.key, required this.api, this.authService});

  @override
  State<ServersScreen> createState() => _ServersScreenState();
}

class _ServersScreenState extends State<ServersScreen> {
  List<RelayServer>? _servers;
  String? _error;

  @override
  void initState() {
    super.initState();
    _load();
  }

  Future<void> _load() async {
    setState(() => _error = null);
    try {
      final servers = await widget.api.fetchServers();
      if (!mounted) return;
      setState(() => _servers = servers);
    } on SessionExpiredException catch (e) {
      if (!mounted) return;
      if (widget.authService != null) {
        navigateToLogin(context, widget.authService!, widget.api);
      } else {
        setState(() => _error = e.message);
      }
    } catch (e) {
      if (!mounted) return;
      setState(() => _error = e is ApiException ? e.message : 'Не удалось загрузить список серверов');
    }
  }

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: const Text('Серверы'),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: RefreshIndicator(
        onRefresh: _load,
        child: _error != null
            ? ListView(
                children: [
                  const SizedBox(height: 60),
                  Center(
                    child: Text(_error!, style: const TextStyle(color: AppColors.danger)),
                  ),
                ],
              )
            : _servers == null
                ? const Center(child: CircularProgressIndicator())
                : _servers!.isEmpty
                    ? ListView(
                        children: const [
                          SizedBox(height: 60),
                          Center(child: Text('Нет доступных серверов')),
                        ],
                      )
                    : ListView.separated(
                        padding: const EdgeInsets.all(18),
                        itemCount: _servers!.length,
                        separatorBuilder: (_, __) => const SizedBox(height: 10),
                        itemBuilder: (context, index) => _ServerTile(server: _servers![index]),
                      ),
      ),
    );
  }
}

class _ServerTile extends StatelessWidget {
  final RelayServer server;
  const _ServerTile({required this.server});

  @override
  Widget build(BuildContext context) {
    return Container(
      padding: const EdgeInsets.all(16),
      decoration: BoxDecoration(
        color: AppColors.surface,
        borderRadius: BorderRadius.circular(18),
        border: Border.all(color: Colors.white.withOpacity(0.08)),
      ),
      child: Row(
        children: [
          Icon(
            server.healthy ? Icons.check_circle_rounded : Icons.error_rounded,
            color: server.healthy ? AppColors.green : AppColors.danger,
          ),
          const SizedBox(width: 14),
          Expanded(
            child: Column(
              crossAxisAlignment: CrossAxisAlignment.start,
              children: [
                Text(server.region, style: Theme.of(context).textTheme.titleMedium),
                const SizedBox(height: 4),
                Text(
                  server.healthy ? 'Доступен' : 'Недоступен',
                  style: Theme.of(context).textTheme.bodySmall,
                ),
              ],
            ),
          ),
          Text(
            '${server.rttMs} ms',
            style: Theme.of(context).textTheme.titleMedium?.copyWith(
                  color: AppColors.green,
                  fontSize: 15,
                ),
          ),
        ],
      ),
    );
  }
}