import 'package:flutter/material.dart';
import '../theme/app_theme.dart';

/// User-level overrides on top of the server-delivered Rule Engine
/// (ТЗ §6: "пользовательские исключения" alongside Domain/CIDR rules).
/// These are sent to the backend as User Rules so the Decision Engine
/// applies them ahead of the default DIRECT/RELAY table.
class ExclusionsScreen extends StatefulWidget {
  final List<String> initial;
  const ExclusionsScreen({super.key, required this.initial});

  @override
  State<ExclusionsScreen> createState() => _ExclusionsScreenState();
}

class _ExclusionsScreenState extends State<ExclusionsScreen> {
  late List<String> _domains = List.of(widget.initial);
  final _inputCtrl = TextEditingController();
  String? _error;

  static final _domainPattern =
      RegExp(r'^(\*\.)?[a-zA-Zа-яА-Я0-9-]+(\.[a-zA-Zа-яА-Я0-9-]+)+$');

  void _add() {
    final value = _inputCtrl.text.trim().toLowerCase();
    if (value.isEmpty) return;

    if (!_domainPattern.hasMatch(value)) {
      setState(() => _error = 'Введите домен, например *.example.com');
      return;
    }
    if (_domains.contains(value)) {
      setState(() => _error = 'Этот домен уже добавлен');
      return;
    }

    setState(() {
      _domains.add(value);
      _inputCtrl.clear();
      _error = null;
    });
  }

  void _remove(String domain) {
    setState(() => _domains.remove(domain));
  }

  @override
  void dispose() {
    _inputCtrl.dispose();
    super.dispose();
  }

  @override
  Widget build(BuildContext context) {
    return WillPopScope(
      onWillPop: () async {
        Navigator.of(context).pop(_domains);
        return false;
      },
      child: Scaffold(
        appBar: AppBar(
          title: const Text('Исключения'),
          backgroundColor: Colors.transparent,
          elevation: 0,
        ),
        body: Column(
          children: [
            Padding(
              padding: const EdgeInsets.all(16),
              child: Column(
                crossAxisAlignment: CrossAxisAlignment.start,
                children: [
                  Text(
                    'Домены из этого списка всегда идут напрямую, '
                    'минуя relay — даже если правило сервера говорит иначе.',
                    style: Theme.of(context).textTheme.bodyMedium,
                  ),
                  const SizedBox(height: 16),
                  Row(
                    children: [
                      Expanded(
                        child: TextField(
                          controller: _inputCtrl,
                          style: const TextStyle(color: AppColors.textPrimary),
                          decoration: const InputDecoration(
                            hintText: '*.mybank.ru',
                          ),
                          onSubmitted: (_) => _add(),
                        ),
                      ),
                      const SizedBox(width: 8),
                      IconButton.filled(
                        onPressed: _add,
                        icon: const Icon(Icons.add),
                        style: IconButton.styleFrom(
                          backgroundColor: AppColors.cyan,
                          foregroundColor: AppColors.bg,
                        ),
                      ),
                    ],
                  ),
                  if (_error != null) ...[
                    const SizedBox(height: 8),
                    Text(_error!,
                        style: const TextStyle(color: AppColors.danger, fontSize: 13)),
                  ],
                ],
              ),
            ),
            const Divider(height: 1),
            Expanded(
              child: _domains.isEmpty
                  ? Center(
                      child: Text(
                        'Список пуст',
                        style: Theme.of(context).textTheme.bodyMedium,
                      ),
                    )
                  : ListView.separated(
                      itemCount: _domains.length,
                      separatorBuilder: (_, __) => const Divider(height: 1),
                      itemBuilder: (context, i) {
                        final domain = _domains[i];
                        return ListTile(
                          title: Text(domain,
                              style: const TextStyle(color: AppColors.textPrimary)),
                          leading: const Icon(Icons.public_off,
                              color: AppColors.textSecondary),
                          trailing: IconButton(
                            icon: const Icon(Icons.close, color: AppColors.textSecondary),
                            onPressed: () => _remove(domain),
                          ),
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
