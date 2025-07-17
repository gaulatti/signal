import { Duration, Stack } from 'aws-cdk-lib';
import { Certificate, ICertificate } from 'aws-cdk-lib/aws-certificatemanager';
import {
  CacheCookieBehavior,
  CacheHeaderBehavior,
  CachePolicy,
  CacheQueryStringBehavior,
  Distribution,
  ErrorResponse,
  OriginAccessIdentity,
  ResponseHeadersPolicy,
  SecurityPolicyProtocol,
  ViewerProtocolPolicy,
} from 'aws-cdk-lib/aws-cloudfront';
import { S3BucketOrigin } from 'aws-cdk-lib/aws-cloudfront-origins';
import { CnameRecord, HostedZone, IHostedZone } from 'aws-cdk-lib/aws-route53';
import { Bucket } from 'aws-cdk-lib/aws-s3';

/**
 * Creates a hosted zone using the provided stack and hosted zone ID.
 *
 * @param stack - The stack object.
 * @returns An object containing the created hosted zone.
 */
const createHostedZone = (stack: Stack) => {
  /**
   * HostedZone
   */
  const hostedZone = HostedZone.fromHostedZoneAttributes(stack, `${stack.stackName}HostedZone`, {
    hostedZoneId: process.env.HOSTED_ZONE_ID!,
    zoneName: process.env.HOSTED_ZONE_NAME!,
  });

  return { hostedZone };
};

/**
 * Builds a zone certificate for the given stack and hosted zone.
 *
 * @param stack - The stack object.
 * @returns The built certificate object.
 */
const createZoneCertificate = (stack: Stack) => {
  /**
   * Certificate
   */
  const certificate = Certificate.fromCertificateArn(stack, `${stack.stackName}Certificate`, process.env.HOSTED_ZONE_CERTIFICATE!);

  return { certificate };
};

/**
 * Creates a CloudFront distribution for the specified stack and S3 bucket.
 *
 * @param stack - The AWS CloudFormation stack.
 * @param bucket - The S3 bucket to be used as the origin source.
 * @param certificate - The SSL certificate to be used by the distribution.
 * @returns The created CloudFront distribution.
 */
const createDistribution = (stack: Stack, s3BucketSource: Bucket, certificate: ICertificate) => {
  /**
   * Represents the CloudFront Origin Access Identity (OAI).
   */
  const originAccessIdentity = new OriginAccessIdentity(stack, `${stack.stackName}DistributionOAI`);

  /**
   * Grants read permissions to the CloudFront Origin Access Identity (OAI).
   */
  s3BucketSource.grantRead(originAccessIdentity);

  /**
   * Creates a cache policy for static assets with a long-term caching strategy.
   *
   * This policy is configured to cache static assets for 1 year with immutable settings.
   * It supports gzip and Brotli encoding for optimized content delivery.
   *
   * - **Cache Policy Name**: Derived from the stack name with a suffix `-StaticAssetsPolicy`.
   * - **TTL Settings**: Default, minimum, and maximum TTL are all set to 365 days.
   * - **Cookie Behavior**: No cookies are included in the cache key.
   * - **Header Behavior**: No headers are included in the cache key.
   * - **Query String Behavior**: No query strings are included in the cache key.
   *
   * @param stack - The CDK stack in which this cache policy is defined.
   * @param stack.stackName - The name of the stack, used to generate the cache policy name.
   */
  const longTermCachePolicy = new CachePolicy(stack, `${stack.stackName}LongTermCachePolicy`, {
    cachePolicyName: `${stack.stackName}LongTermCachePolicy`,
    defaultTtl: Duration.days(365),
    minTtl: Duration.days(365),
    maxTtl: Duration.days(365),
    enableAcceptEncodingGzip: true,
    enableAcceptEncodingBrotli: true,
    cookieBehavior: CacheCookieBehavior.none(),
    headerBehavior: CacheHeaderBehavior.none(),
    queryStringBehavior: CacheQueryStringBehavior.none(),
  });

  /**
   * Creates a new `ResponseHeadersPolicy` for the given stack with custom headers behavior.
   *
   * This policy is configured to add a `CacheControl` header with the value
   * `public, max-age=31536000, immutable`, which ensures that static assets are cached
   * by clients for one year and are treated as immutable. The `override` property is set
   * to `true`, meaning this header will replace any existing `Cache-Control` header.
   *
   * @param stack - The CDK stack in which the `ResponseHeadersPolicy` is defined.
   * @param stack.stackName - The name of the stack, used to generate unique identifiers.
   * @returns A `ResponseHeadersPolicy` instance with the specified custom headers behavior.
   */
  const responseHeadersPolicy = new ResponseHeadersPolicy(stack, `${stack.stackName}ResponseHeadersPolicy`, {
    responseHeadersPolicyName: `${stack.stackName}ResponseHeadersPolicy`,
    customHeadersBehavior: {
      customHeaders: [
        {
          header: 'Cache-Control',
          value: 'public, max-age=31536000, immutable',
          override: true,
        },
      ],
    },
  });

  /**
   * Represents the CloudFront distribution to serve the React Assets.
   */
  const distribution = new Distribution(stack, `${stack.stackName}Distribution`, {
    defaultBehavior: {
      origin: S3BucketOrigin.withOriginAccessIdentity(s3BucketSource, { originAccessIdentity }),
      viewerProtocolPolicy: ViewerProtocolPolicy.REDIRECT_TO_HTTPS,
      cachePolicy: longTermCachePolicy,
      responseHeadersPolicy: responseHeadersPolicy,
    },
    defaultRootObject: 'index.html',
    domainNames: [`signal.${process.env.HOSTED_ZONE_NAME}`],
    certificate,
    minimumProtocolVersion: SecurityPolicyProtocol.TLS_V1_2_2021,
    errorResponses: [
      {
        httpStatus: 404,
        responseHttpStatus: 200,
        responsePagePath: '/index.html',
        ttl: Duration.seconds(0),
      } as ErrorResponse,
    ],
  });

  return { distribution };
};

/**
 * Creates a CNAME record for the frontend.
 *
 * @param stack - The AWS CloudFormation stack.
 * @param zone - The hosted zone where the CNAME record will be created.
 * @param distribution - The distribution associated with the CNAME record.
 * @returns An object containing the frontend CNAME record.
 */
const createCNAME = (stack: Stack, zone: IHostedZone, distribution: Distribution) => {
  const record = new CnameRecord(stack, `${stack.stackName}FrontendCNAME`, {
    recordName: 'signal',
    zone,
    domainName: distribution.distributionDomainName,
  });

  return { record };
};

export { createCNAME, createDistribution, createHostedZone, createZoneCertificate };
