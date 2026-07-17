# Lessons Learned

## 1. Mismatched DOM Element IDs (Frontend Crashing)
- **Correction:** The user reported that clicking other games did not update the UI or products.
- **Pattern:** The JavaScript code was updated to change a title element using `document.getElementById('form-game-title')`, but the actual HTML element ID was `gh-title`. This caused a silent JS TypeError: "Cannot read properties of null (setting 'textContent')", which broke the entire click handler execution.
- **Rule for Future:** Always verify that DOM IDs used in JavaScript files exist in the corresponding HTML files. Do not assume old/default IDs are present. Run a quick check/script to match IDs when editing UI event handlers.

## 2. Overly Aggressive Categorization Regex
- **Correction:** Products like "1X Weekly Card" were rendered as just "1" with a diamond icon.
- **Pattern:** The categorizer function (`categorize`) initially classified any product without numbers in its name OR with `diamond === 0` as a subscription. However, it did not handle products where a number was part of a pass name (e.g. "1X Weekly Card" has a number "1" but is still a subscription).
- **Rule for Future:** Do not rely purely on number extraction to determine if an item is a standard top-up vs a subscription. Use explicit keyword lists (e.g. 'pass', 'weekly', 'starlight') to match subscriptions, and fall back to price tiers for standard items.

## 3. Order ID Concurrency Collision
- **Pattern:** Using pure seconds-based timestamps (`DML<timestamp_seconds>`) for order numbers has a high collision risk if two users purchase at the same second.
- **Rule for Future:** Always append a random suffix (e.g. `rand.Intn(1000)`) or use millisecond-level precision to ensure uniqueness of transactional IDs.
