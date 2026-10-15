import os
import requests
import telebot
from telebot import types

TOKEN = "8800467868:AAEZ4XWZRWAzX6j2A4L6F9gvuPovbZ2-jx8"
ADMIN_ID = 2011675494

bot = telebot.TeleBot(TOKEN)

GROUPS_FILE = "bot_database.txt"
VIOLATIONS_LOG_FILE = "violations_log.txt"

def save_group(chat_id, chat_title):
    try:
        if chat_id >= 0:
            return
        groups = load_groups()
        groups[str(chat_id)] = chat_title
        with open(GROUPS_FILE, "w", encoding="utf-8") as f:
            for g_id, g_title in groups.items():
                f.write(f"{g_id}:{g_title}\n")
    except Exception as e:
        print(f"Error saving group: {e}")

def load_groups():
    groups = {}
    if os.path.exists(GROUPS_FILE):
        try:
            with open(GROUPS_FILE, "r", encoding="utf-8") as f:
                for line in f:
                    if ":" in line:
                        parts = line.strip().split(":", 1)
                        if len(parts) == 2 and parts[0].startswith("-"):
                            groups[parts[0]] = parts[1]
        except Exception as e:
            print(f"Error loading groups: {e}")
    return groups

def remove_group(chat_id):
    try:
        groups = load_groups()
        if str(chat_id) in groups:
            del groups[str(chat_id)]
            with open(GROUPS_FILE, "w", encoding="utf-8") as f:
                for g_id, g_title in groups.items():
                    f.write(f"{g_id}:{g_title}\n")
    except Exception as e:
        print(f"Error removing group: {e}")

def get_violation_count():
    if not os.path.exists(VIOLATIONS_LOG_FILE):
        return 0
    count = 0
    with open(VIOLATIONS_LOG_FILE, "r", encoding="utf-8") as f:
        for _ in f:
            count += 1
    return count

def save_violation_record(chat_id, user_id, user_name, chat_title, reason):
    try:
        with open(VIOLATIONS_LOG_FILE, "a", encoding="utf-8") as f:
            f.write(f"{chat_id}|{user_id}|{user_name}|{chat_title}|{reason}\n")
    except Exception as e:
        print(f"Error logging violation: {e}")

def get_last_violation():
    if not os.path.exists(VIOLATIONS_LOG_FILE):
        return None
    last_record = None
    try:
        with open(VIOLATIONS_LOG_FILE, "r", encoding="utf-8") as f:
            for line in f:
                parts = line.strip().split("|")
                if len(parts) == 5:
                    last_record = {
                        "chat_id": parts[0],
                        "user_id": parts[1],
                        "user_name": parts[2],
                        "chat_title": parts[3],
                        "reason": parts[4]
                    }
    except Exception as e:
        print(f"Error reading violation log: {e}")
    return last_record

STRINGS = {
    'ar': {
        'welcome': "🦅 **أهلاً بك في نظام صقر الكردي الذكي!**\n\nمرحباً يا **{}**، أنا جاهز لحماية المجموعات.",
        'contact_dev': "👨‍💻 التواصل مع المطور",
        'lang_btn': "🌐 اللغة: العربية 🇸🇦",
        'violations_btn': "📋 سجل المخالفات",
        'leave_menu_btn': "🚪 خروج البوت من المجموعة",
        'settings_btn': "⚙️ إعدادات المجموعات"
    },
    'en': {
        'welcome': "🦅 **Welcome to Saqar Kurdish Smart System!**\n\nHello **{}**, I am ready to protect groups.",
        'contact_dev': "👨‍💻 Contact Developer",
        'lang_btn': "🌐 Language: English 🇬🇧",
        'violations_btn': "📋 Violations Log",
        'leave_menu_btn': "🚪 Bot Leave Group",
        'settings_btn': "⚙️ Groups Settings"
    }
}

user_languages = {}

def get_lang(uid):
    return user_languages.get(uid, 'ar')

def t(uid, key, *args):
    lang = get_lang(uid)
    txt = STRINGS.get(lang, STRINGS['ar']).get(key, key)
    return txt.format(*args) if args else txt

@bot.my_chat_member_handler()
def handle_chat_member(event):
    chat = event.chat
    new_status = event.new_chat_member.status
    old_status = event.old_chat_member.status
    
    if chat.type in ['group', 'supergroup']:
        if new_status in ['member', 'administrator'] and old_status not in ['member', 'administrator']:
            markup = types.InlineKeyboardMarkup(row_width=2)
            accept_btn = types.InlineKeyboardButton("✅ موافقة", callback_data=f"accept_group_{chat.id}")
            reject_btn = types.InlineKeyboardButton("❌ رفض", callback_data=f"reject_group_{chat.id}")
            
            if chat.username:
                chat_link_btn = types.InlineKeyboardButton("💬 التواصل مع المجموعة", url=f"https://t.me/{chat.username}")
                markup.row(accept_btn, reject_btn)
                markup.row(chat_link_btn)
            else:
                markup.row(accept_btn, reject_btn)
            
            try:
                bot.send_message(
                    ADMIN_ID,
                    f"⚠️ **طلب إضافة بوت جديد لمجموعة!**\n\n"
                    f"📌 اسم المجموعة: `{chat.title}`\n"
                    f"🆔 آيدي المجموعة: `{chat.id}`\n\n"
                    f"هل تود السماح للبوت بالدخول لهذه المجموعة؟",
                    parse_mode="Markdown",
                    reply_markup=markup
                )
            except Exception as e:
                print(f"Error notifying admin: {e}")
                
        elif new_status in ['left', 'kicked']:
            remove_group(chat.id)

@bot.message_handler(func=lambda message: message.chat.type in ['group', 'supergroup'])
def auto_track_groups(message):
    groups = load_groups()
    if str(message.chat.id) in groups:
        save_group(message.chat.id, message.chat.title)

# --- استجابة فورية لكلمة "دليل" داخل المجموعة للمطور فقط ---
@bot.message_handler(func=lambda message: message.text and "دليل" in message.text and message.chat.type in ['group', 'supergroup'] and message.from_user.id == ADMIN_ID)
def show_violation_proof_in_group(message):
    record = get_last_violation()
    if not record:
        bot.reply_to(message, "⚠️ لا توجد مخالفات مسجلة في السجل حتى الآن.")
        return
    
    v_num = get_violation_count()
    proof_text = (
        f"📌 **المخالفة رقم ({v_num})**\n\n"
        f"👥 **المجموعة:** {record['chat_title']}\n"
        f"👤 **العضو:** {record['user_name']}\n"
        f"🆔 **المعرف:** `{record['user_id']}`\n"
        f"🚨 **السبب:** {record['reason']}\n"
        f"👮 **المسؤول:** نظام الحماية الآلي (صقر الكردي)"
    )
    
    markup = types.InlineKeyboardMarkup(row_width=2)
    unban_btn = types.InlineKeyboardButton("🔓 فك الحظر", callback_data=f"unban_{record['chat_id']}_{record['user_id']}")
    contact_btn = types.InlineKeyboardButton("💬 التواصل مع العضو", url=f"tg://user?id={record['user_id']}")
    markup.row(unban_btn, contact_btn)
    
    bot.reply_to(message, proof_text, parse_mode="Markdown", reply_markup=markup)

# --- نظام الحماية الفعالة للصور، الفيديوهات، واللقطات المتحركة ---
@bot.message_handler(content_types=['photo', 'animation', 'video'], func=lambda message: message.chat.type in ['group', 'supergroup'])
def handle_media_protection(message):
    chat_id = message.chat.id
    groups = load_groups()
    
    if str(chat_id) not in groups:
        return

    try:
        user_status = bot.get_chat_member(chat_id, message.from_user.id).status
        if user_status in ['administrator', 'creator'] or message.from_user.id == ADMIN_ID:
            return
    except:
        pass

    user_id = message.from_user.id
    user_name = message.from_user.first_name
    chat_title = message.chat.title
    
    reason = "نشر صور ضيقة غير لائقة."
    if message.animation:
        reason = "نشر لقطة متحركة (GIF) غير لائقة."
    elif message.video:
        reason = "نشر فيديو غير لائق أو مخالف."
    elif message.photo:
        reason = "نشر صورة شبه إباحي."

    save_violation_record(chat_id, user_id, user_name, chat_title, reason)
    v_num = get_violation_count()

    try:
        proof_text = (
            f"📌 **المخالفة رقم ({v_num})**\n\n"
            f"👥 **المجموعة:** {chat_title}\n"
            f"👤 **العضو:** {user_name}\n"
            f"🆔 **المعرف:** `{user_id}`\n"
            f"🚨 **السبب:** {reason}\n"
            f"👮 **المسؤول:** نظام الحماية الآلي (صقر الكردي)"
        )
        markup = types.InlineKeyboardMarkup(row_width=2)
        unban_btn = types.InlineKeyboardButton("🔓 فك الحظر", callback_data=f"unban_{chat_id}_{user_id}")
        contact_btn = types.InlineKeyboardButton("💬 التواصل مع العضو", url=f"tg://user?id={user_id}")
        markup.row(unban_btn, contact_btn)

        # إرسال الدليل لخاص المطور تلقائياً
        if message.photo:
            bot.send_photo(ADMIN_ID, message.photo[-1].file_id, caption=proof_text, parse_mode="Markdown", reply_markup=markup)
        else:
            bot.send_message(ADMIN_ID, proof_text, parse_mode="Markdown", reply_markup=markup)

        bot.delete_message(chat_id, message.message_id)
        bot.ban_chat_member(chat_id, user_id)

        alert_text = (
            "🚨 **تنبيه حماية (صقر الكردي)**\n\n"
            f"👤 **العضو:** {user_name}\n"
            f"🆔 **المعرف:** `{user_id}`\n"
            f"❌ **السبب:** {reason}\n"
            "🚫 **الإجراء:** تم الحذف والحظر وتوثيق المخالفة في السجل."
        )
        bot.send_message(chat_id, alert_text, parse_mode="Markdown")
    except Exception as e:
        print(f"Error in media protection: {e}")

@bot.message_handler(commands=['start', 'settings'])
def start_cmd(message):
    uid = message.from_user.id
    name = message.from_user.first_name
    
    markup = types.InlineKeyboardMarkup(row_width=2)
    dev_btn = types.InlineKeyboardButton(t(uid, 'contact_dev'), url=f"tg://user?id={ADMIN_ID}")
    lang_btn = types.InlineKeyboardButton(t(uid, 'lang_btn'), callback_data="toggle_language")
    markup.row(dev_btn, lang_btn)
    
    if message.chat.type == 'private' and uid == ADMIN_ID:
        viol_btn = types.InlineKeyboardButton(t(uid, 'violations_btn'), callback_data="select_group_violations")
        leave_menu_btn = types.InlineKeyboardButton(t(uid, 'leave_menu_btn'), callback_data="open_leave_section")
        settings_btn = types.InlineKeyboardButton(t(uid, 'settings_btn'), callback_data="open_settings_section")
        markup.row(viol_btn, leave_menu_btn)
        markup.row(settings_btn)
        
    bot.reply_to(message, t(uid, 'welcome', name), parse_mode="Markdown", reply_markup=markup)

@bot.callback_query_handler(func=lambda call: call.data in ["toggle_language", "select_group_violations", "open_leave_section", "open_settings_section", "back_to_main"] or call.data.startswith("leave_chat_") or call.data.startswith("group_settings_") or call.data.startswith("accept_group_") or call.data.startswith("reject_group_") or call.data.startswith("violations_chat_") or call.data.startswith("unban_"))
def start_callbacks(call):
    uid = call.from_user.id
    
    # --- معالجة زر فك الحظر الحقيقي وتحديث الواجهة إلى "تم الفك" ---
    if call.data.startswith("unban_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
        
        parts = call.data.split("_")
        if len(parts) >= 3:
            target_chat_id = parts[1]
            target_user_id = parts[2]
            
            try:
                # فك الحظر الفعلي عن العضو في تلك المجموعة
                bot.unban_chat_member(int(target_chat_id), int(target_user_id), only_if_banned=True)
                
                # تحديث شكل الزر في الرسالة ليصبح "تم الفك" ومعطلاً
                old_markup = call.message.reply_markup
                new_markup = types.InlineKeyboardMarkup(row_width=2)
                
                unban_done_btn = types.InlineKeyboardButton("✅ تم الفك", callback_data="noop")
                
                # استخراج رابط التواصل القديم إن وجد
                contact_url = f"tg://user?id={target_user_id}"
                if old_markup:
                    for row in old_markup.keyboard:
                        for btn in row:
                            if "التواصل" in btn.text:
                                contact_url = btn.url
                                
                contact_btn = types.InlineKeyboardButton("💬 التواصل مع العضو", url=contact_url)
                new_markup.row(unban_done_btn, contact_btn)
                
                if call.message.photo:
                    bot.edit_message_caption(
                        chat_id=call.message.chat.id,
                        message_id=call.message.message_id,
                        caption=call.message.caption + "\n\n🔓 **حالة الحظر:** تم فك الحظر عن العضو بنجاح.",
                        parse_mode="Markdown",
                        reply_markup=new_markup
                    )
                else:
                    bot.edit_message_text(
                        text=call.message.text + "\n\n🔓 **حالة الحظر:** تم فك الحظر عن العضو بنجاح.",
                        chat_id=call.message.chat.id,
                        message_id=call.message.message_id,
                        parse_mode="Markdown",
                        reply_markup=new_markup
                    )
                
                bot.answer_callback_query(call.id, "✅ تم فك الحظر عن العضو في المجموعة بنجاح!", show_alert=True)
            except Exception as e:
                bot.answer_callback_query(call.id, f"❌ حدث خطأ أثناء فك الحظر: {e}", show_alert=True)
        return

    if call.data == "toggle_language":
        user_languages[uid] = 'en' if get_lang(uid) == 'ar' else 'ar'
        bot.answer_callback_query(call.id, "✅ تم تغيير اللغة بنجاح.", show_alert=True)
        
        name = call.from_user.first_name
        markup = types.InlineKeyboardMarkup(row_width=2)
        dev_btn = types.InlineKeyboardButton(t(uid, 'contact_dev'), url=f"tg://user?id={ADMIN_ID}")
        lang_btn = types.InlineKeyboardButton(t(uid, 'lang_btn'), callback_data="toggle_language")
        markup.row(dev_btn, lang_btn)
        
        if call.message.chat.type == 'private' and uid == ADMIN_ID:
            viol_btn = types.InlineKeyboardButton(t(uid, 'violations_btn'), callback_data="select_group_violations")
            leave_menu_btn = types.InlineKeyboardButton(t(uid, 'leave_menu_btn'), callback_data="open_leave_section")
            settings_btn = types.InlineKeyboardButton(t(uid, 'settings_btn'), callback_data="open_settings_section")
            markup.row(viol_btn, leave_menu_btn)
            markup.row(settings_btn)
            
        try:
            bot.edit_message_text(t(uid, 'welcome', name), call.message.chat.id, call.message.message_id, parse_mode="Markdown", reply_markup=markup)
        except:
            pass
            
    elif call.data == "select_group_violations":
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
        
        groups = load_groups()
        if not groups:
            bot.answer_callback_query(call.id, "⚠️ لا توجد مجموعات مسجلة حالياً.", show_alert=True)
            return
            
        markup = types.InlineKeyboardMarkup(row_width=1)
        for g_id, g_title in groups.items():
            markup.add(types.InlineKeyboardButton(f"📁 {g_title}", callback_data=f"violations_chat_{g_id}"))
            
        markup.add(types.InlineKeyboardButton("🔙 رجوع", callback_data="back_to_main"))
            
        bot.answer_callback_query(call.id, "📋 جارِ تحميل سجل المخالفات...")
        try:
            bot.edit_message_text(
                "📋 **سجل المخالفات:**\nاختر المجموعة لعرض مخالفاتها المسجلة:", 
                call.message.chat.id,
                call.message.message_id,
                parse_mode="Markdown",
                reply_markup=markup
            )
        except Exception:
            pass

    elif call.data.startswith("violations_chat_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("violations_chat_", "")
        groups = load_groups()
        g_title = groups.get(chat_id, "المجموعة")
        
        record = get_last_violation()
        markup = types.InlineKeyboardMarkup(row_width=1)
        markup.add(types.InlineKeyboardButton("🔙 رجوع", callback_data="select_group_violations"))
        
        if not record:
            text = f"📋 **سجل مخالفات مجموعة ({g_title}):**\n\n✅ لا توجد مخالفات مسجلة حتى الآن."
        else:
            text = (
                f"📌 **آخر مخالفة مسجلة في ({g_title}):**\n\n"
                f"👤 **العضو:** {record['user_name']}\n"
                f"🆔 **المعرف:** `{record['user_id']}`\n"
                f"🚨 **السبب:** {record['reason']}\n"
                f"👮 **المسؤول:** نظام الحماية الآلي (صقر الكردي)"
            )
        
        bot.answer_callback_query(call.id, "📋 تم جلب سجل المخالفات.")
        try:
            bot.edit_message_text(text, call.message.chat.id, call.message.message_id, parse_mode="Markdown", reply_markup=markup)
        except Exception:
            pass

    elif call.data == "open_settings_section":
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
        
        groups = load_groups()
        if not groups:
            bot.answer_callback_query(call.id, "⚠️ لا توجد مجموعات مسجلة حالياً.", show_alert=True)
            return
            
        markup = types.InlineKeyboardMarkup(row_width=1)
        for g_id, g_title in groups.items():
            markup.add(types.InlineKeyboardButton(f"⚙️ {g_title}", callback_data=f"group_settings_{g_id}"))
            
        markup.add(types.InlineKeyboardButton("🔙 رجوع", callback_data="back_to_main"))
            
        bot.answer_callback_query(call.id, "⚙️ تم فتح إعدادات المجموعات...")
        try:
            bot.edit_message_text(
                "⚙️ **إعدادات المجموعات:**\nاختر المجموعة للتحكم بإعداداتها وعقوباتها:",
                call.message.chat.id,
                call.message.message_id,
                parse_mode="Markdown",
                reply_markup=markup
            )
        except Exception:
            pass

    elif call.data.startswith("group_settings_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("group_settings_", "")
        groups = load_groups()
        g_title = groups.get(chat_id, "المجموعة")
        
        markup = types.InlineKeyboardMarkup(row_width=1)
        markup.add(types.InlineKeyboardButton("🛡️ تفعيل/تعطيل الحماية (قريباً)", callback_data=f"toggle_prot_{chat_id}"))
        markup.add(types.InlineKeyboardButton("⚙️ تعديل العقوبات (قريباً)", callback_data=f"edit_pen_{chat_id}"))
        markup.add(types.InlineKeyboardButton("🔙 رجوع لقائمة المجموعات", callback_data="open_settings_section"))
        
        bot.answer_callback_query(call.id, f"⚙️ مجموعة: {g_title}")
        try:
            bot.edit_message_text(
                f"⚙️ **لوحة إعدادات المجموعة:**\n"
                f"📌 اسم المجموعة: `{g_title}`\n"
                f"🆔 الآيدي: `{chat_id}`\n\n"
                f"اختر الإجراء المطلوب:",
                call.message.chat.id,
                call.message.message_id,
                parse_mode="Markdown",
                reply_markup=markup
            )
        except Exception:
            pass

    elif call.data == "open_leave_section":
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
        
        groups = load_groups()
        if not groups:
            bot.answer_callback_query(call.id, "⚠️ لا توجد مجموعات حقيقية مسجلة حالياً.", show_alert=True)
            return
            
        markup = types.InlineKeyboardMarkup(row_width=1)
        for g_id, g_title in groups.items():
            markup.add(types.InlineKeyboardButton(f"{g_title}", callback_data=f"leave_chat_{g_id}"))
            
        markup.add(types.InlineKeyboardButton("🔙 رجوع", callback_data="back_to_main"))
            
        bot.answer_callback_query(call.id, "📂 تم فتح قسم المجموعات...")
        try:
            bot.edit_message_text(
                "📂 **اختر المجموعة التي تريد مغادرتها:**", 
                call.message.chat.id,
                call.message.message_id,
                parse_mode="Markdown",
                reply_markup=markup
            )
        except Exception:
            pass

    elif call.data == "back_to_main":
        if uid != ADMIN_ID:
            return
        name = call.from_user.first_name
        markup = types.InlineKeyboardMarkup(row_width=2)
        dev_btn = types.InlineKeyboardButton(t(uid, 'contact_dev'), url=f"tg://user?id={ADMIN_ID}")
        lang_btn = types.InlineKeyboardButton(t(uid, 'lang_btn'), callback_data="toggle_language")
        markup.row(dev_btn, lang_btn)
        
        if call.message.chat.type == 'private' and uid == ADMIN_ID:
            viol_btn = types.InlineKeyboardButton(t(uid, 'violations_btn'), callback_data="select_group_violations")
            leave_menu_btn = types.InlineKeyboardButton(t(uid, 'leave_menu_btn'), callback_data="open_leave_section")
            settings_btn = types.InlineKeyboardButton(t(uid, 'settings_btn'), callback_data="open_settings_section")
            markup.row(viol_btn, leave_menu_btn)
            markup.row(settings_btn)
            
        try:
            bot.edit_message_text(t(uid, 'welcome', name), call.message.chat.id, call.message.message_id, parse_mode="Markdown", reply_markup=markup)
        except:
            pass

    elif call.data.startswith("leave_chat_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("leave_chat_", "")
        try:
            bot.leave_chat(int(chat_id))
            remove_group(chat_id)
            bot.answer_callback_query(call.id, "✅ تم خروج البوت من المجموعة بنجاح!", show_alert=True)
            bot.edit_message_text("✅ تم تنفيذ أمر الخروج وحذف المجموعة من القائمة بنجاح.", call.message.chat.id, call.message.message_id, parse_mode="Markdown")
        except Exception as e:
            bot.answer_callback_query(call.id, f"❌ حدث خطأ: {e}", show_alert=True)

    elif call.data.startswith("accept_group_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الإجراء للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("accept_group_", "")
        try:
            chat_info = bot.get_chat(int(chat_id))
            save_group(chat_info.id, chat_info.title)
            
            announcement = (
                "🦅 **تمت الموافقة الرسمية!**\n\n"
                "تمت موافقة إدخال وتفعيل البوت لهذه المجموعة من قبل **المطور صقر الكردي** ✨\n\n"
                "🛡️ *النظام جاهز الآن لحماية وإدارة المجموعة.*"
            )
            bot.send_message(int(chat_id), announcement, parse_mode="Markdown")
            
            bot.answer_callback_query(call.id, "✅ تم قبول الطلب وإرسال الإشعار للمجموعة.")
            bot.edit_message_text(f"✅ **تمت الموافقة على مجموعة `{chat_info.title}` وإرسال إشعار التفعيل بنجاح.**", call.message.chat.id, call.message.message_id, parse_mode="Markdown")
        except Exception as e:
            bot.answer_callback_query(call.id, "❌ حدث خطأ: تأكد من وجود البوت داخل المجموعة.", show_alert=True)

    elif call.data.startswith("reject_group_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الإجراء للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("reject_group_", "")
        try:
            bot.leave_chat(int(chat_id))
            remove_group(chat_id)
            bot.answer_callback_query(call.id, "❌ تم رفض الطلب وغادرت المجموعة.")
            bot.edit_message_text("❌ **تم رفض إضافة البوت وغادر المجموعة بنجاح.**", call.message.chat.id, call.message.message_id, parse_mode="Markdown")
        except Exception as e:
            bot.answer_callback_query(call.id, f"❌ حدث خطأ: {e}", show_alert=True)

if __name__ == "__main__":
    print("Bot running with active unban action and UI state update...")
    bot.infinity_polling(timeout=10, long_polling_timeout=5)
