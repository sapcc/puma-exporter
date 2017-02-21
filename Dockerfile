FROM scratch

COPY bin/puma_exporter /puma_exporter

EXPOSE 9325

ENTRYPOINT [ "/puma_exporter" ]
