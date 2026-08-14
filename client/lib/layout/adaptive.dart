import 'package:flutter/material.dart';

/// Wide window: side rail + constrained content (Windows desktop).
const kWideLayoutBreakpoint = 900.0;

bool isWideLayout(BuildContext context) =>
    MediaQuery.sizeOf(context).width >= kWideLayoutBreakpoint;
