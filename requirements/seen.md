
## 📌 Feature: **Message Seen Details (Channels + DM)**

---

## 🎯 Mục tiêu

Cung cấp khả năng **xem danh sách người đã xem một tin nhắn cụ thể**, bao gồm **thời điểm xem**, áp dụng cho:

* Public channels
* Private channels
* **Direct Messages (1–1)**
* **Group Direct Messages**

Giải quyết nhu cầu:

* Biết chính xác **ai đã xem – khi nào**
* Tránh tranh cãi “chưa thấy / chưa đọc”
* Hỗ trợ debug và audit trong môi trường enterprise

---

## 🧭 Phạm vi áp dụng

### ✔ Supported

* Channel messages
* DM (1–1)
* Group DM

### ❌ Not supported

* System messages
* Ephemeral / bot-only messages (configurable)

---

## 🖱️ UX / UI Behavior

### 1️⃣ Entry point (mọi loại message)

* Trên mỗi message:

  * Right-click / more actions (⋮) → **“Seen by”**
  * Hoặc click **Seen indicator** (✓✓ / text)

---

### 2️⃣ Seen Indicator theo loại hội thoại

#### 🟢 Channel

* Hiển thị:

  ```
  Seen by Anthony, Quang, +10
  ```
* Click → mở danh sách đầy đủ

#### 🔵 DM (1–1)

* Không hiển thị list dài
* Chỉ hiển thị:

  * `Seen`
  * `Seen at 16:36`
* Hover để xem timestamp

#### 🟣 Group DM

* Hiển thị:

  ```
  Seen by 3/5
  ```
* Click → xem chi tiết từng người + thời gian

---

### 3️⃣ Seen List Modal / Popover

Mỗi item gồm:

* Avatar
* Display name
* Username
* Seen time:

  * Absolute: `Today at 16:36`
  * Relative: `2 minutes ago`

Sắp xếp:

```
Latest seen → Oldest
```

---

## ⏱️ Định nghĩa “đã xem”

Một user được coi là **đã xem message** khi:

* Conversation đang active
* Message nằm trong viewport
* Thời gian hiển thị ≥ X ms (configurable, mặc định 500ms)
* Không tính:

  * Background tab
  * Minimized app
  * Fast scroll

---

## 🔁 Realtime Behavior

* Seen updates qua websocket
* Append-only:

  * Chỉ ghi **first seen**
* UI update ngay khi event tới

---



### 🧾 Logging (debug & audit)

```json
{
  "event": "message_seen",
  "message_id": "...",
  "conversation_type": "dm",
  "conversation_id": "...",
  "user_id": "...",
  "seen_at": "...",
  "client": "mobile",
  "viewport_ms": 812
}
```



## 🔐 Privacy & Control

### DM-specific rules

* **DM 1–1**:

  * Chỉ 2 người thấy trạng thái
  * Không expose seen list API
* **Group DM**:

  * Chỉ members trong DM thấy list
* Admin controls:

  * Disable seen for DM
  * Disable logging chi tiết

---

## 🧪 Acceptance Criteria

* [ ] Seen hoạt động cho Channel + DM
* [ ] DM 1–1 hiển thị “Seen at …”
* [ ] Group DM hiển thị seen list đúng
* [ ] Realtime update chính xác
* [ ] Không leak data ngoài conversation
* [ ] Performance ổn với DM active lớn

---

## 🧠 Edge Cases (bổ sung cho DM)

* User mở DM nhưng không scroll tới message
* User xem trên mobile trước, web sau
* DM bị mute
* User block nhau (DM 1–1)
* User rời group DM sau khi seen
