

## 📌 Feature: Global **Unread All** (Mattermost)

### 🎯 Mục tiêu

Hiện tại, sidebar chỉ hiển thị **unread theo từng team**, gây khó khăn cho người dùng khi theo dõi tất cả tin nhắn chưa đọc.
Tính năng **Unread All** nhằm cung cấp một **view tập trung** cho toàn bộ tin nhắn chưa đọc, cập nhật realtime, giúp người dùng truy cập nhanh và không bỏ lỡ thông tin quan trọng.

---

### 🧭 Phạm vi & UX tổng thể

* Thu thập **tất cả tin nhắn chưa đọc** từ:

  * Tất cả teams
  * Tất cả channels (public / private / direct / group)
* Không phụ thuộc vào team đang active
* Cập nhật **realtime** theo websocket / events

---

### 🖱️ UI / Placement

* Thêm **nút “Unread All”** tại **Header**
* Vị trí:

  * Bên cạnh **Recent Mentions**
* Trạng thái:

  * Hiển thị badge số lượng unread
  * Active / Inactive state rõ ràng

---

### 📥 Nội dung hiển thị (Unread All View)

Mỗi item unread cần bao gồm:

* Team name
* Channel name
* Sender
* Message preview (1–2 dòng)
* Timestamp
* Unread indicator (bold / dot)

Danh sách được sắp xếp theo:

```
Newest unread message → Oldest
```

---

### 🔁 Realtime Behavior

* Tự động thêm item mới khi có message chưa đọc
* Tự động remove item khi:

  * User đọc message
  * Channel được mark as read
* Badge counter cập nhật ngay lập tức

---

### 🚀 Navigation / Jump behavior

* Khi click vào một unread item:

  * Chuyển đúng **team**
  * Mở đúng **channel**
  * Scroll & focus vào **vị trí chính xác của message**
  * Highlight message trong vài giây (optional)

---

### ⚙️ Logic & Data

* Unread được xác định dựa trên:

  * `last_viewed_at`
  * `last_post_at`
* Không duplicate message nếu cùng channel
* Handle edge cases:

  * Channel bị archive
  * User bị remove khỏi team/channel
  * Message bị delete

---

### 🧪 Acceptance Criteria

* [ ] User có thể xem tất cả unread messages ở một nơi duy nhất
* [ ] Badge unread count chính xác & realtime
* [ ] Click vào message jump đúng vị trí
* [ ] Không phụ thuộc team đang active
* [ ] Performance ổn với số lượng unread lớn

---

### 🧠 Optional Enhancements (Nice-to-have)

* Filter:

  * By team
  * By channel type (DM / Channel)
* Keyboard shortcut:

  * `Cmd/Ctrl + Shift + U`
* Mark all as read
* Pin important unread

---
