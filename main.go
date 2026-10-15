import os
import requests
import telebot
from telebot import types

TOKEN = "8800467868:AAEZ4XWZRWAzX6j2A4L6F9gvuPovbZ2-jx8"
ADMIN_ID = 2011675494

bot = telebot.TeleBot(TOKEN)

GROUPS_FILE = "bot_database.txt"

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

STRINGS = {
    'ar': {
        'welcome': "🦅 **أهلاً بك في نظام صقر الكردي الذكي!**\n\nمرحباً يا **{}**، أنا جاهز لحماية المجموعات.",
        'contact_dev': "👨‍💻 التواصل مع المطور",
        'lang_btn': "🌐 اللغة: العربية 🇸🇦",
        'violations_btn': "📋 سجل المخالفات",
        'leave_menu_btn': "🚪 خروج البوت من المجموعة"
    },
    'en': {
        'welcome': "🦅 **Welcome to Saqar Kurdish Smart System!**\n\nHello **{}**, I am ready to protect groups.",
        'contact_dev': "👨‍💻 Contact Developer",
        'lang_btn': "🌐 Language: English 🇬🇧",
        'violations_btn': "📋 Violations Log",
        'leave_menu_btn': "🚪 Bot Leave Group"
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

@bot.message_handler(commands=['start'])
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
        markup.row(viol_btn, leave_menu_btn)
        
    bot.reply_to(message, t(uid, 'welcome', name), parse_mode="Markdown", reply_markup=markup)

@bot.callback_query_handler(func=lambda call: call.data in ["toggle_language", "select_group_violations", "open_leave_section", "back_to_main"] or call.data.startswith("leave_chat_") or call.data.startswith("accept_group_") or call.data.startswith("reject_group_") or call.data.startswith("violations_chat_"))
def start_callbacks(call):
    uid = call.from_user.id
    
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
            markup.row(viol_btn, leave_menu_btn)
            
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
            
        bot.answer_callback_query(call.id, "📋 جارِ تحميل قائمة المجموعات...")
        bot.send_message(
            call.message.chat.id, 
            "📋 **سجل المخالفات:**\nاختر المجموعة لعرض المخالفات الخاصة بها:", 
            parse_mode="Markdown",
            reply_markup=markup
        )

    elif call.data.startswith("violations_chat_"):
        if uid != ADMIN_ID:
            bot.answer_callback_query(call.id, "❌ هذا الزر للمطور فقط.", show_alert=True)
            return
            
        chat_id = call.data.replace("violations_chat_", "")
        groups = load_groups()
        g_title = groups.get(chat_id, "المجموعة")
        
        markup = types.InlineKeyboardMarkup(row_width=1)
        markup.add(types.InlineKeyboardButton("🔙 رجوع", callback_data="select_group_violations"))
        
        bot.answer_callback_query(call.id, "📋 تم جلب سجل المخالفات.")
        bot.send_message(
            call.message.chat.id,
            f"📋 **سجل مخالفات مجموعة ({g_title}):**\n\n"
            f"✅ لا توجد مخالفات مسجلة حتى الآن في هذه المجموعة.",
            parse_mode="Markdown",
            reply_markup=markup
        )

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
        bot.send_message(
            call.message.chat.id, 
            "📂 **اختر المجموعة التي تريد مغادرتها:**", 
            parse_mode="Markdown",
            reply_markup=markup
        )

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
            markup.row(viol_btn, leave_menu_btn)
            
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
    print("Bot running cleanly without invisible characters...")
    bot.infinity_polling(timeout=10, long_polling_timeout=5)
