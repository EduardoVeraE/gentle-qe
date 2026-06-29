---
name: selenium-e2e-java
description: "Trigger: Selenium WebDriver, Java E2E tests, Page Object Model, JUnit 5, AssertJ, Allure, Maven. Author and maintain Selenium/Java end-to-end suites."
license: Apache-2.0
metadata:
  author: gentle-qe
---

# Selenium WebDriver E2E Testing (Java)

Toolkit for end-to-end browser test automation using **Selenium WebDriver with Java 21+**, JUnit 5, AssertJ soft assertions, the Page Object Model, and Allure reporting. This is the Java/Selenium counterpart to the `playwright-e2e-testing` skill — use it when the project's stack is JVM-based rather than Node/TypeScript.

> **Activation:** Triggered when working with Selenium WebDriver, Java-based E2E tests, Page Object Model in Java, JUnit 5 test structure, AssertJ assertions, WebDriver factories, or Maven test projects.

## When to Use This Skill

- **Write E2E tests** for user flows on projects with a Java/JVM stack
- **Build a Page Object Model** with `BasePage`/`BaseTest` and a thread-safe `WebDriverFactory`
- **Set up explicit waits** correctly (never `Thread.sleep`) using `WebDriverWait` + `ExpectedConditions`
- **Structure JUnit 5 suites** with proper annotations, naming, and Allure reporting
- **Apply AssertJ soft assertions** for richer, non-fail-fast verification
- **Enable parallel execution and retries** via JUnit 5 platform config
- **Wire CI/CD** for the Selenium suite (GitHub Actions)

## Prerequisites

| Requirement     | Details                                              |
| --------------- | --------------------------------------------------- |
| JDK             | Java 21+ (modern language features used throughout) |
| Build tool      | Maven (standard `src/test/java` layout)             |
| Selenium        | `selenium-java` 4.x                                 |
| Test framework  | JUnit 5 (Jupiter)                                   |
| Assertions      | AssertJ (soft assertions)                           |
| Reporting       | Allure                                              |

## Core Principles (Test Constitution — Java/Selenium)

**MUST DO**

1. **Page Object Model** — encapsulate locators and interactions in page classes; specs never touch raw `WebDriver` calls.
2. **Locator priority** — prefer stable, semantic locators; declare them as `By` constants in the page object.
3. **Explicit waits only** — `WebDriverWait` + `ExpectedConditions`; tests must be deterministic.
4. **External test data** — never hardcode strings, IDs, URLs, or credentials.
5. **`@Step` annotations** — wrap logical groupings for Allure traceability.
6. **Thread-safe driver** — obtain `WebDriver` from a factory (`ThreadLocal`) to support parallel runs.

**WON'T DO**

1. **NEVER** use `Thread.sleep()` or other hard waits.
2. **NEVER** use fragile absolute XPath.
3. **NEVER** hardcode data or share mutable driver state across threads.
4. **NEVER** skip running the suite to verify it passes.

## References

| Document                                                            | Content                                                                 |
| ------------------------------------------------------------------- | ----------------------------------------------------------------------- |
| [Selenium + Java Guidelines](./references/selenium-java-guidelines.md) | Full guide: tech stack, locator strategy, POM, explicit waits, JUnit 5 annotations, AssertJ soft assertions, file organization, BasePage/BaseTest, WebDriver factory, complete page/test examples, retry mechanism, parallel execution, and CI/CD integration. |

## Verification

- Suite runs green via `mvn test` (or the project's Maven goal).
- No `Thread.sleep` / hard waits anywhere (grep the suite).
- All page objects extend `BasePage`; all tests extend `BaseTest`.
- Allure report generates with `@Step`-annotated traceability.
