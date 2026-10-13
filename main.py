import os
import cv2
import requests
from pyrogram import Client, filters
from pyrogram.types import InlineKeyboardMarkup, InlineKeyboardButton, Message, CallbackQuery

API_ID = 21226626
API_HASH = "ea1a0c2fa9587a9df2a3325056efe110"
BOT_TOKEN = "ضع_التوكن_هنا"
OWNER_ID = 2011675494

SIGHTENGINE_USER = "ضع_api_user_هنا"
SIGHTENGINE_SECRET = "ضع_api_secret_هنا"

app = Client("SaqrBot", api_id=API_ID, api_hash=API_HASH, bot_token=BOT_TOKEN)

def check_image_nsfw(file_path: str) -> bool:
    try:
        params = {
            'models': 'nudity-2.0,wad',
            'api_user': SIGHTENGINE_USER,
            'api_secret': SIGHTENGINE_SECRET
        }
        with open(file_path, 'rb') as f:
            files = {'media': f}
            response = requests.post('https://api.sightengine.com/1.0/check.json', files=files, data=params, timeout=15)
            output = response.json()
            
            if output.get('status') == 'success':
                nudity = output.get('nudity', {})
                raw = nudity.get('raw', 0)
                partial = nudity.get('partial', 0)
                sexual_activity = nudity.get('sexual_activity', 0)
                sexual_display = nudity.get('sexual_display', 0)
                erotica = nudity.get('erotica', 0)
                
                if max(raw, partial, sexual_activity, sexual_display, erotica) > 0.5:
                    return True
    except Exception as e:
        print(f"خطأ في الاتصال بخدمة الفحص: {e}")
    return False

def check_video_nsfw(video_path: str) -> bool:
    cap = cv2.VideoCapture(video_path)
    frame_count = 0
    is_nsfw = False
    try:
        while cap.isOpened():
            ret, frame = cap.read()
            if not ret:
                break
            if frame_count % 30 == 0:
                temp_frame_path = f"temp_frame_{frame_count}.jpg"
                cv2.imwrite(temp_frame_path, frame)
                if check_image_nsfw(temp_frame_path):
                    is_nsfw = True
                    if os.path.exists(temp_frame_path):
                        os.remove(temp_frame_path)
                    break
                if os.path.exists(temp_frame_path):
                    os.remove(temp_frame_path)
            frame_count += 1
    except Exception as e:
        print(f"خطأ أثناء فحص الفيديو: {e}")
    finally:
        cap.release()
    return is_nsfw

@app.on_message(filters.new_chat_members)
async def new_chat_member_handler(client, message: Message):
    for member in message.new_chat_members:
        if member.id == client.me.id:
            chat = message.chat
            user = message.from_user
            text = (
                f"🚨 **طلب إضافة بوت جديد للمجموعة!**\n\n"
                f"🏷 **اسم المجموعة:** {chat.title}\n"
                f"🆔 **آيدي المجموعة:** `{chat.id}`\n"
                f"👤 **المستخدم الذي أضافني:** {user.mention} (`{user.id}`)\n\n"
                f"اختر الإجراء المناسب للمجموعة:"
            )
            kb = InlineKeyboardMarkup([
                [
                    InlineKeyboardButton("موافق ✅", callback_data=f"accept_{chat.id}"),
                    InlineKeyboardButton("رفض ومغادرة ❌", callback_data=f"reject_{chat.id}")
                ]
            ])
            try:
                await client.send_message(OWNER_ID, text, reply_markup=kb)
            except Exception as e:
                print(f"خطأ في إرسال التنبيه للمطور: {e}")

@app.on_callback_query()
async def callback_handler(client, callback_query: CallbackQuery):
    data = callback_query.data
    if callback_query.from_user.id != OWNER_ID:
        await callback_query.answer("هذا الزر ليس مخصصاً لك!", show_alert=True)
        return

    action, chat_id = data.split("_")
    chat_id = int(chat_id)

    if action == "accept":
        await callback_query.message.edit_text(f"✅ **تم قبول المجموعة بنجاح وتفعيل البوت فيها.**")
        try:
            await client.send_message(chat_id, "🦅 تم تفعيل بوت صقر الكردي في هذه المجموعة بنجاح بواسطة المطور.")
        except:
            pass
    elif action == "reject":
        await callback_query.message.edit_text(f"❌ **تم رفض المجموعة ومغادرتها.**")
        try:
            await client.leave_chat(chat_id)
        except Exception as e:
            print(f"خطأ عند مغادرة المجموعة: {e}")

@app.on_message(filters.group & (filters.photo | filters.document))
async def handle_media_photos(client, message: Message):
    file_path = await message.download()
    try:
        if check_image_nsfw(file_path):
            user = message.from_user
            await message.delete()
            await client.ban_chat_member(message.chat.id, user.id)
            await message.reply_text(
                f"🚨 **حظر فوري بسبب مخالفة القوانين!**\n"
                f"👤 **العضو:** {user.mention}\n"
                f"⚠️ **السبب:** نشر صور أو لقطات محتوى غير لائق / إباحي."
            )
    except Exception as e:
        print(f"خطأ معالجة الصورة: {e}")
    finally:
        if file_path and os.path.exists(file_path):
            os.remove(file_path)

@app.on_message(filters.group & filters.video)
async def handle_media_videos(client, message: Message):
    file_path = await message.download()
    try:
        if check_video_nsfw(file_path):
            user = message.from_user
            await message.delete()
            await client.ban_chat_member(message.chat.id, user.id)
            await message.reply_text(
                f"🚨 **حظر فوري بسبب مخالفة القوانين!**\n"
                f"👤 **العضو:** {user.mention}\n"
                f"⚠️ **السبب:** نشر مقطع فيديو يحتوي على محتوى غير لائق."
            )
    except Exception as e:
        print(f"خطأ معالجة الفيديو: {e}")
    finally:
        if file_path and os.path.exists(file_path):
            os.remove(file_path)

if __name__ == "__main__":
    print("🦅 بوت صقر الكردي بحلته الجديدة يعمل الآن...")
    app.run()
