# Open Questions

> Questions requiring product/architect decision

---

| ID | Question | Priority | Decision Needed | Status |
|----|----------|----------|-----------------|--------|
| Q-001 | Какую Go-библиотеку Hysteria2 использовать в go_core? (sing-box, official hy2 client, other?) | P0 | Tech choice | Open |
| Q-002 | Есть ли sandbox ключи ЮKassa для live-тестирования? | P0 | Credentials | Open |
| Q-003 | Subscription cancel: немедленная отмена или «stop auto-renewal»? (ADR-010) | P1 | Product | Open |
| Q-004 | Production domain: api.streampass.com или другой? | P1 | Infrastructure | Open |
| Q-005 | Нужен ли refresh token auto-rotation на клиенте? | P2 | Feature | Open |
| Q-006 | Exclusions: синхронизировать с backend или оставить local-only? | P2 | Feature | Resolved — sync (BL-014) |
| Q-007 | Prometheus/Grafana: self-hosted или managed (Grafana Cloud)? | P2 | Infrastructure | Open |
| Q-008 | Android release keystore: где хранить, кто управляет? | P1 | Security | Partial — Gradle+key.properties ready; store JKS offline |
| Q-009 | Telemetry retention period: сколько дней хранить events? | P3 | Privacy | Open |
| Q-010 | Beta testing: сколько пользователей, как recruit? | P2 | Product | Open |
| Q-011 | iOS priority: когда начинать после Android MVP? | P2 | Roadmap | Open |
| Q-012 | connection_config: шифровать at rest в PostgreSQL? | P2 | Security | Open |

---

## How to Resolve

1. Product questions (Q-003, Q-006, Q-010) → Product Owner
2. Tech questions (Q-001, Q-005, Q-007) → Tech Lead / Architect
3. Security questions (Q-008, Q-012) → Security review
4. Infrastructure (Q-004) → DevOps

After decision: update relevant docs + close question here.
