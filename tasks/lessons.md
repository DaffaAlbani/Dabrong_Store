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

## 4. Hard-Disabled Submit Buttons Without Feedback
- **Correction:** The user reported "saya gabisa klik bayar" (cannot click pay / buy button).
- **Pattern:** `checkStep2Next()` set `btn.disabled = true` if input fields (e.g. WhatsApp or package selection) were incomplete or didn't match a strict regex (`/^0\d{9,12}$/`). Native HTML `disabled` attribute prevented any click events from firing, leaving the user with zero visual feedback or toast notifications explaining why they couldn't submit. Furthermore, strict regex rejected valid Indonesian phone numbers starting with `628` or `+628`.
- **Rule for Future:** Avoid hard-disabling form submission buttons silently. Keep buttons interactive or provide explicit Toast / Tooltip feedback on click detailing missing fields (User ID, Server ID, Package, valid phone number format).

## 5. No Unrequested International UI Sections on Local Web Store
- **Correction:** The user requested "kok web nya ada international chekout?? gaperlu itu, aku cm mau yg indo aja, rubah sprti awal lg".
- **Pattern:** Added a 3-tier international Paddle checkout section directly onto the public homepage of an Indonesian game top-up web store.
- **Rule for Future:** Never inject unrequested global/international pricing sections or foreign checkout widgets onto the main consumer-facing store UI. Keep the user interface clean, focused, and tailored strictly to Indonesian payment methods (QRIS, Bank Transfer BCA, Account Saldo).

## 6. Avoid Adding Unrequested Content Bars or Extra Sections
- **Correction:** The user requested "Menampilkan 4 indikator kepercayaaan utama di bawah Hero Section... gaperlu ini".
- **Pattern:** Added an extra 4-metric trust stats bar between the Hero section and Marketplace section during visual polish.
- **Rule for Future:** When polishing UI aesthetics, refine existing cards, typography, and hover effects without adding unrequested new content blocks or marketing banners unless explicitly asked by the user.
