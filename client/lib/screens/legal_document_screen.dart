import 'package:flutter/material.dart';
import '../theme/app_theme.dart';
import '../legal/legal_docs.dart';

/// Simple scrollable legal document (E01 → BL-054).
class LegalDocumentScreen extends StatelessWidget {
  final String title;
  final String body;

  const LegalDocumentScreen({
    super.key,
    required this.title,
    required this.body,
  });

  factory LegalDocumentScreen.terms({Key? key}) => LegalDocumentScreen(
        key: key,
        title: LegalDocs.termsTitle,
        body: LegalDocs.termsBody,
      );

  factory LegalDocumentScreen.privacy({Key? key}) => LegalDocumentScreen(
        key: key,
        title: LegalDocs.privacyTitle,
        body: LegalDocs.privacyBody,
      );

  factory LegalDocumentScreen.refund({Key? key}) => LegalDocumentScreen(
        key: key,
        title: LegalDocs.refundTitle,
        body: LegalDocs.refundBody,
      );

  factory LegalDocumentScreen.support({Key? key}) => LegalDocumentScreen(
        key: key,
        title: LegalDocs.supportTitle,
        body: LegalDocs.supportBody,
      );

  @override
  Widget build(BuildContext context) {
    return Scaffold(
      appBar: AppBar(
        title: Text(title),
        backgroundColor: Colors.transparent,
        elevation: 0,
      ),
      body: SingleChildScrollView(
        padding: const EdgeInsets.fromLTRB(24, 8, 24, 32),
        child: Text(
          body.trim(),
          style: const TextStyle(
            color: AppColors.textPrimary,
            height: 1.45,
            fontSize: 14,
          ),
        ),
      ),
    );
  }
}
