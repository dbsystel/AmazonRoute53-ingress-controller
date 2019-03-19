FROM alpine:3.8

RUN apk add --no-cache curl

RUN addgroup -S kube-operator && adduser -S -g kube-operator kube-operator

USER kube-operator

COPY ./bin/AmazonRoute53-ingress-controller .

ENTRYPOINT ["./AmazonRoute53-ingress-controller"]
