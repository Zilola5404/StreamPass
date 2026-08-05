import 'package:flutter/material.dart';
import 'package:flutter/services.dart';

import '../theme/app_theme.dart';

class InstalledAppInfo {
  final String packageName;
  final String label;

  const InstalledAppInfo({required this.packageName, required this.label});

  factory InstalledAppInfo.fromMap(Map<dynamic, dynamic> map) {
    return InstalledAppInfo(
      packageName: map['packageName'] as String? ?? '',
      label: map['label'] as String? ?? '',
    );
  }
}

/// Picker for Android packages that must bypass VpnService entirely
/// (TRANSPORT_VPN = false). Complements domain exclusions.
class AppBypassScreen extends StatefulWidget {
  final List<String> initialSelected;

  const AppBypassScreen({super.key, required this.initialSelected});

  @override
  State<AppBypassScreen> createState() => _AppBypassScreenState();
}

class _AppBypassScreenState extends State<AppBypassScreen> {
  static const _channel = MethodChannel('streampass/installed_apps');

  final _selected = <String>{};
  List<InstalledAppInfo> _apps = const [];
  bool _loading = true;
  String? _error;
  String _query = '';

  @override
  void initState() {
    super.initState();
    _selected.addAll(widget.initialSelected);
    _load();
  }

  Future<void> _load() async {
    try {
      final raw = await _channel.invokeMethod<List<dynamic>>('listLaunchable');
      final apps = (raw ?? const [])
          .whereType<Map>()
          .map((m) => InstalledAppInfo.fromMap(m))
          .where((a) => a.packageName.isNotEmpty)
          .toList();
      if (!mounted) return;
      setState(() {
        _apps = apps;
        _loading = false;
        _error = null;
      });
    } on MissingPluginException {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = 'Список приложений доступен только на Android';
      });
    } catch (e) {
      if (!mounted) return;
      setState(() {
        _loading = false;
        _error = '$e';
      });
    }
  }

  List<InstalledAppInfo> get _filtered {
    final q = _query.trim().toLowerCase();
    if (q.isEmpty) return _apps;
    return _apps
        .where(
          (a) =>
              a.label.toLowerCase().contains(q) ||
              a.packageName.toLowerCase().contains(q),
        )
        .toList();
  }

  @override
  Widget build(BuildContext context) {
    return WillPopScope(
      onWillPop: () async {
        Navigator.of(context).pop(_selected.toList()..sort());
        return false;
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Приложения без VPN'),
          backgroundColor: Colors.transparent,
          elevation: 0,
          actions: [
            TextButton(
              onPressed: () =>
                  Navigator.of(context).pop(_selected.toList()..sort()),
              child: const Text('Готово'),
            ),
          ],
        ),
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.fromLTRB(16, 8, 16, 8),
              child: Text(
                'Выбранные приложения полностью обходят StreamPass на уровне ОС '
                '(нужно для Госуслуг, банков и ФНС). Изменения применятся при следующем подключении.',
                style: Theme.of(context).textTheme.bodySmall,
              ),
            ),
            Padding(
              padding: const EdgeInsets.symmetric(horizontal: 16),
              child: TextField(
                decoration: const InputDecoration(
                  hintText: 'Поиск по названию или package',
                  prefixIcon: Icon(Icons.search),
                ),
                onChanged: (v) => setState(() => _query = v),
              ),
            ),
            const SizedBox(height: 8),
            if (_loading)
              const Expanded(child: Center(child: CircularProgressIndicator()))
            else if (_error != null)
              Expanded(
                child: Center(
                  child: Text(_error!, style: const TextStyle(color: AppColors.danger)),
                ),
              )
            else
              Expanded(
                child: ListView.builder(
                  itemCount: _filtered.length,
                  itemBuilder: (context, index) {
                    final app = _filtered[index];
                    final checked = _selected.contains(app.packageName);
                    return CheckboxListTile(
                      value: checked,
                      activeColor: AppColors.cyan,
                      title: Text(app.label),
                      subtitle: Text(
                        app.packageName,
                        style: Theme.of(context).textTheme.bodySmall,
                      ),
                      onChanged: (v) {
                        setState(() {
                          if (v == true) {
                            _selected.add(app.packageName);
                          } else {
                            _selected.remove(app.packageName);
                          }
                        });
                      },
                    );
                  },
                ),
              ),
          ],
        ),
      ),
    );
  }
}
