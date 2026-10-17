#pragma once

#include <QString>

class ProxySettings {
public:
    ProxySettings();

    bool hasPendingRestore() const;
    bool apply(int socks_port, QString& error);
    bool restore(QString& error);

private:
    QString backup_path_;
};
