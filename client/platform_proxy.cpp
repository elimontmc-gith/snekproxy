#include "platform_proxy.h"

#include <QDir>
#include <QFile>
#include <QFileDevice>
#include <QProcess>
#include <QSettings>
#include <QStandardPaths>
#include <QStringList>
#include <QVariant>

#ifdef Q_OS_WIN
#include <windows.h>
#include <wininet.h>
#endif

namespace {

constexpr auto kBackendKey = "proxy/backend";
constexpr auto kPendingKey = "proxy/pending";

#ifdef Q_OS_LINUX
constexpr auto kGnomeSchema = "org.gnome.system.proxy";
constexpr auto kGnomeSocksSchema = "org.gnome.system.proxy.socks";

bool run_gsettings(const QStringList& arguments, QString& output, QString& error) {
    QProcess process;
    process.start("gsettings", arguments);
    if (!process.waitForStarted(3000)) {
        error = "Could not start gsettings. This Linux desktop proxy integration requires GNOME "
                "and the gsettings utility.";
        return false;
    }
    if (!process.waitForFinished(5000)) {
        process.kill();
        process.waitForFinished();
        error = "gsettings timed out while updating the desktop proxy.";
        return false;
    }
    output = QString::fromLocal8Bit(process.readAllStandardOutput()).trimmed();
    if (process.exitStatus() != QProcess::NormalExit || process.exitCode() != 0) {
        error = QString::fromLocal8Bit(process.readAllStandardError()).trimmed();
        if (error.isEmpty()) {
            error = "gsettings could not access the GNOME proxy settings.";
        }
        return false;
    }
    return true;
}

bool get_gsettings(const QString& schema, const QString& key, QString& value, QString& error) {
    QString output;
    if (!run_gsettings({"get", schema, key}, output, error)) {
        return false;
    }
    value = output;
    return true;
}

bool set_gsettings(const QString& schema, const QString& key, const QString& value,
                   QString& error) {
    QString output;
    return run_gsettings({"set", schema, key, value}, output, error);
}

bool save_file_settings(const QString& path, const QString& backend,
                        const QStringList& keys, const QStringList& values, QString& error) {
    QSettings backup(path, QSettings::IniFormat);
    if (backup.contains(kPendingKey) && backup.value(kPendingKey).toBool()) {
        error = "A previous proxy change still needs restoration. Use Disable Tunnel or restart "
                "the app to restore it before connecting again.";
        return false;
    }
    backup.clear();
    backup.setValue(kBackendKey, backend);
    backup.setValue(kPendingKey, true);
    for (qsizetype i = 0; i < keys.size(); ++i) {
        backup.setValue("proxy/values/" + keys[i], values[i]);
    }
    backup.sync();
    if (backup.status() != QSettings::NoError) {
        error = "Could not save the previous proxy settings for safe restoration.";
        return false;
    }
    QFile::setPermissions(path, QFileDevice::ReadOwner | QFileDevice::WriteOwner);
    return true;
}
#endif

#ifdef Q_OS_WIN
constexpr auto kInternetSettingsPath =
    "HKEY_CURRENT_USER\\Software\\Microsoft\\Windows\\CurrentVersion\\Internet Settings";
const QStringList kWindowsKeys{
    "ProxyEnable",
    "ProxyServer",
    "AutoConfigURL",
    "AutoDetect",
};

void notify_windows_proxy_change() {
    InternetSetOptionW(nullptr, INTERNET_OPTION_SETTINGS_CHANGED, nullptr, 0);
    InternetSetOptionW(nullptr, INTERNET_OPTION_REFRESH, nullptr, 0);
}

bool set_windows_proxy_value(const QString& key, const QVariant& value, bool present,
                             QString& error) {
    QSettings settings(kInternetSettingsPath, QSettings::NativeFormat);
    if (present) {
        settings.setValue(key, value);
    } else {
        settings.remove(key);
    }
    settings.sync();
    if (settings.status() != QSettings::NoError) {
        error = "Could not update the current user's Windows Internet proxy settings.";
        return false;
    }
    return true;
}
#endif

}  // namespace

ProxySettings::ProxySettings() {
    auto directory = QStandardPaths::writableLocation(QStandardPaths::AppConfigLocation);
    QDir().mkpath(directory);
    backup_path_ = directory + "/proxy-backup.ini";
}

bool ProxySettings::hasPendingRestore() const {
    QSettings backup(backup_path_, QSettings::IniFormat);
    return backup.value(kPendingKey, false).toBool();
}

bool ProxySettings::apply(int socks_port, QString& error) {
    if (hasPendingRestore()) {
        error = "Previous proxy settings have not been restored yet. Disable the previous tunnel "
                "or restart the app to recover them.";
        return false;
    }

#ifdef Q_OS_LINUX
    const QStringList keys{"mode", "socks/host", "socks/port"};
    QStringList originals;
    for (const auto& item : keys) {
        const auto parts = item.split('/');
        const auto schema = parts.size() == 1 ? QString(kGnomeSchema) : QString(kGnomeSocksSchema);
        const auto key = parts.last();
        QString value;
        if (!get_gsettings(schema, key, value, error)) {
            return false;
        }
        originals.append(value);
    }
    if (!save_file_settings(backup_path_, "gnome", keys, originals, error)) {
        return false;
    }

    const bool applied =
        set_gsettings(kGnomeSchema, "mode", "'manual'", error) &&
        set_gsettings(kGnomeSocksSchema, "host", "'127.0.0.1'", error) &&
        set_gsettings(kGnomeSocksSchema, "port", QString::number(socks_port), error);
    if (!applied) {
        QString restore_error;
        if (!restore(restore_error)) {
            error += "\nThe original settings could not be restored: " + restore_error;
        }
        return false;
    }
    return true;
#elif defined(Q_OS_WIN)
    QSettings registry(kInternetSettingsPath, QSettings::NativeFormat);
    QSettings backup(backup_path_, QSettings::IniFormat);
    if (backup.value(kPendingKey, false).toBool()) {
        error = "Previous proxy settings have not been restored yet. Disable the previous tunnel "
                "or restart the app to restore them before connecting again.";
        return false;
    }
    backup.clear();
    backup.setValue(kBackendKey, "windows");
    backup.setValue(kPendingKey, true);
    for (const auto& key : kWindowsKeys) {
        const bool exists = registry.contains(key);
        backup.setValue("proxy/present/" + key, exists);
        if (exists) {
            backup.setValue("proxy/values/" + key, registry.value(key));
        }
    }
    backup.sync();
    if (backup.status() != QSettings::NoError) {
        error = "Could not save the previous proxy settings for safe restoration.";
        return false;
    }
    QFile::setPermissions(backup_path_, QFileDevice::ReadOwner | QFileDevice::WriteOwner);

    if (!set_windows_proxy_value("AutoConfigURL", QString(), true, error) ||
        !set_windows_proxy_value("AutoDetect", 0, true, error) ||
        !set_windows_proxy_value("ProxyServer",
                                 QString("socks=127.0.0.1:%1").arg(socks_port), true, error) ||
        !set_windows_proxy_value("ProxyEnable", 1, true, error)) {
        QString restore_error;
        if (!restore(restore_error)) {
            error += "\nThe original settings could not be restored: " + restore_error;
        }
        return false;
    }
    notify_windows_proxy_change();
    return true;
#else
    Q_UNUSED(socks_port);
    error = "Automatic system proxy setup is not supported on this operating system.";
    return false;
#endif
}

bool ProxySettings::restore(QString& error) {
    QSettings backup(backup_path_, QSettings::IniFormat);
    if (!backup.value(kPendingKey, false).toBool()) {
        return true;
    }

    const auto backend = backup.value(kBackendKey).toString();
#ifdef Q_OS_LINUX
    if (backend != "gnome") {
        error = "Saved proxy settings use an unknown backend; refusing to alter desktop settings.";
        return false;
    }
    const QStringList items{"mode", "socks/host", "socks/port"};
    for (auto item = items.crbegin(); item != items.crend(); ++item) {
        const auto parts = item->split('/');
        const auto schema = parts.size() == 1 ? QString(kGnomeSchema) : QString(kGnomeSocksSchema);
        const auto key = parts.last();
        if (!set_gsettings(schema, key, backup.value("proxy/values/" + *item).toString(), error)) {
            return false;
        }
    }
#elif defined(Q_OS_WIN)
    if (backend != "windows") {
        error = "Saved proxy settings use an unknown backend; refusing to alter Windows settings.";
        return false;
    }
    const QStringList restore_order{"AutoConfigURL", "AutoDetect", "ProxyServer", "ProxyEnable"};
    for (const auto& key : restore_order) {
        const bool present = backup.value("proxy/present/" + key, false).toBool();
        const QVariant value = backup.value("proxy/values/" + key);
        if (!set_windows_proxy_value(key, value, present, error)) {
            return false;
        }
    }
    notify_windows_proxy_change();
#else
    Q_UNUSED(backend);
    error = "Automatic proxy restoration is not supported on this operating system.";
    return false;
#endif

    backup.clear();
    backup.sync();
    if (backup.status() != QSettings::NoError) {
        error = "Proxy settings were restored, but the recovery record could not be cleared.";
        return false;
    }
    return true;
}
